package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"pinpeople/internal/domain"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient implements the RedisClient interface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(append([]interface{}{ctx}, toInterfaceSlice(keys)...)...)
	return args.Get(0).(*redis.IntCmd)
}

// Helper function to convert []string to []interface{}
func toInterfaceSlice(slice []string) []interface{} {
	interfaceSlice := make([]interface{}, len(slice))
	for i, v := range slice {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

// MockEmployeeRepository implements the EmployeeRepository interface
type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) FindAllPaginated(page, pageSize int) ([]*domain.Employee, int, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*domain.Employee), args.Int(1), args.Error(2)
}

func (m *MockEmployeeRepository) Create(employee *domain.Employee) (*domain.Employee, error) {
	args := m.Called(employee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Update(employee *domain.Employee) error {
	args := m.Called(employee)
	return args.Error(0)
}

func (m *MockEmployeeRepository) FindByID(id int) (*domain.Employee, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindAll() ([]*domain.Employee, error) {
	args := m.Called()
	return args.Get(0).([]*domain.Employee), args.Error(1)
}

// Helper functions to create Redis command results
func newStringCmd(val string, err error) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background(), "GET")
	cmd.SetVal(val)
	if err != nil {
		cmd.SetErr(err)
	}
	return cmd
}

func newStatusCmd(val string, err error) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(context.Background(), "SET")
	cmd.SetVal(val)
	if err != nil {
		cmd.SetErr(err)
	}
	return cmd
}

func newIntCmd(val int, err error) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background(), "DEL")
	cmd.SetVal(int64(val))
	if err != nil {
		cmd.SetErr(err)
	}
	return cmd
}

// Helper functions to create pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestGetEmployeesHandler(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	// Dados de teste
	employees := []*domain.Employee{
		{ID: 1, Name: "John Doe"},
		{ID: 2, Name: "Jane Smith"},
	}
	totalCount := 2
	page := 1
	pageSize := 10
	totalPages := 1

	// Configurando o mock para Redis GET que retorna redis.Nil (cache miss)
	cacheKey := fmt.Sprintf("%s:page:%d:size:%d", employeesCacheKey, page, pageSize)
	mockRedis.On("Get", mock.Anything, cacheKey).Return(newStringCmd("", redis.Nil))

	// Configurando o mock para buscar do repositório
	mockRepo.On("FindAllPaginated", page, pageSize).Return(employees, totalCount, nil)

	// Configurando o mock para Redis SET
	response := struct {
		Employees  []*domain.Employee `json:"employees"`
		TotalCount int                `json:"totalCount"`
		Page       int                `json:"page"`
		PageSize   int                `json:"pageSize"`
		TotalPages int                `json:"totalPages"`
	}{
		Employees:  employees,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	jsonData, _ := json.Marshal(response)
	mockRedis.On("Set", mock.Anything, cacheKey, jsonData, 5*time.Minute).Return(newStatusCmd("OK", nil))

	// Criando o handler
	handler := GetEmployeesHandler(mockRepo, mockRedis)

	// Criando a requisição
	req, err := http.NewRequest("GET", "/employees?page=1&pageSize=10", nil)
	assert.NoError(t, err)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verificando o corpo da resposta
	var respBody struct {
		Employees  []*domain.Employee `json:"employees"`
		TotalCount int                `json:"totalCount"`
		Page       int                `json:"page"`
		PageSize   int                `json:"pageSize"`
		TotalPages int                `json:"totalPages"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, employees, respBody.Employees)
	assert.Equal(t, totalCount, respBody.TotalCount)
	assert.Equal(t, page, respBody.Page)
	assert.Equal(t, pageSize, respBody.PageSize)
	assert.Equal(t, totalPages, respBody.TotalPages)

	// Verificando as expectativas dos mocks
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateEmployeeHandler(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	// Dados de teste
	newEmployee := &domain.Employee{
		ID:           0, // ID será atribuído pelo repositório
		Name:         "Alice Johnson",
		Position:     stringPtr("Engineer"),
		DepartmentID: intPtr(1), // Ajuste para DepartmentID
	}

	createdEmployee := &domain.Employee{
		ID:           1,
		Name:         "Alice Johnson",
		Position:     stringPtr("Engineer"),
		DepartmentID: intPtr(1),
	}

	// Configurando o mock para criar o funcionário
	// Use mock.AnythingOfType("*domain.Employee") para evitar problemas com valores de ponteiros
	mockRepo.On("Create", mock.AnythingOfType("*domain.Employee")).Return(createdEmployee, nil)

	// Configurando o mock para invalidar o cache
	mockRedis.On("Del", mock.Anything, employeesCacheKey).Return(newIntCmd(1, nil))

	// Criando o handler
	handler := CreateEmployeeHandler(mockRepo, mockRedis)

	// Criando a requisição
	employeeJSON, _ := json.Marshal(newEmployee)
	req, err := http.NewRequest("POST", "/employees", bytes.NewBuffer(employeeJSON))
	assert.NoError(t, err)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verificando o corpo da resposta
	var respBody domain.Employee
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, createdEmployee.ID, respBody.ID)
	assert.Equal(t, createdEmployee.Name, respBody.Name)
	assert.Equal(t, *createdEmployee.Position, *respBody.Position)
	assert.Equal(t, *createdEmployee.DepartmentID, *respBody.DepartmentID)

	// Verificando as expectativas dos mocks
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestDeleteEmployeeHandler(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	employeeID := 1

	// Configurando o mock para deletar o funcionário
	mockRepo.On("Delete", employeeID).Return(nil)

	// Configurando o mock para invalidar o cache
	mockRedis.On("Del", mock.Anything, employeesCacheKey).Return(newIntCmd(1, nil))

	// Criando o handler
	handler := DeleteEmployeeHandler(mockRepo, mockRedis)

	// Criando a requisição
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/employees/%d", employeeID), nil)
	assert.NoError(t, err)

	// Definindo as variáveis de URL
	vars := map[string]string{
		"id": strconv.Itoa(employeeID),
	}
	req = mux.SetURLVars(req, vars)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verificando as expectativas dos mocks
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUpdateEmployeeHandler(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	employeeID := 1

	// Dados de teste
	updateData := &domain.Employee{
		Name:     "Alice Johnson Updated",
		Position: stringPtr("Senior Engineer"),
	}

	updatedEmployee := &domain.Employee{
		ID:       employeeID,
		Name:     "Alice Johnson Updated",
		Position: stringPtr("Senior Engineer"),
	}

	// Configurando o mock para atualizar o funcionário
	// Use mock.AnythingOfType("*domain.Employee") para evitar problemas com valores de ponteiros
	mockRepo.On("Update", mock.AnythingOfType("*domain.Employee")).Return(nil)

	// Configurando o mock para invalidar o cache
	mockRedis.On("Del", mock.Anything, employeesCacheKey).Return(newIntCmd(1, nil))

	// Criando o handler
	handler := UpdateEmployeeHandler(mockRepo, mockRedis)

	// Criando a requisição
	employeeJSON, _ := json.Marshal(updateData)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/employees/%d", employeeID), bytes.NewBuffer(employeeJSON))
	assert.NoError(t, err)

	// Definindo as variáveis de URL
	vars := map[string]string{
		"id": strconv.Itoa(employeeID),
	}
	req = mux.SetURLVars(req, vars)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verificando o corpo da resposta
	var respBody domain.Employee
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, updatedEmployee.ID, respBody.ID)
	assert.Equal(t, updatedEmployee.Name, respBody.Name)
	assert.Equal(t, *updatedEmployee.Position, *respBody.Position)

	// Verificando as expectativas dos mocks
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetEmployeeByIDHandler(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)

	employeeID := 1

	// Dados de teste
	employee := &domain.Employee{
		ID:       employeeID,
		Name:     "John Doe",
		Position: stringPtr("Developer"),
	}

	// Configurando o mock para buscar o funcionário por ID
	mockRepo.On("FindByID", employeeID).Return(employee, nil)

	// Criando o handler
	handler := GetEmployeeByIDHandler(mockRepo)

	// Criando a requisição
	req, err := http.NewRequest("GET", fmt.Sprintf("/employees/%d", employeeID), nil)
	assert.NoError(t, err)

	// Definindo as variáveis de URL
	vars := map[string]string{
		"id": strconv.Itoa(employeeID),
	}
	req = mux.SetURLVars(req, vars)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verificando o corpo da resposta
	var respBody domain.Employee
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, employee.ID, respBody.ID)
	assert.Equal(t, employee.Name, respBody.Name)
	assert.Equal(t, *employee.Position, *respBody.Position)

	// Verificando as expectativas dos mocks
	mockRepo.AssertExpectations(t)
}

func TestGetEmployeesHandlerCacheHit(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	cachedData := `{"employees":[{"id":1,"name":"John Doe"}],"totalCount":1,"page":1,"pageSize":10,"totalPages":1}`
	cacheKey := fmt.Sprintf("%s:page:%d:size:%d", employeesCacheKey, 1, 10)

	mockRedis.On("Get", mock.Anything, cacheKey).Return(newStringCmd(cachedData, nil))

	handler := GetEmployeesHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("GET", "/employees?page=1&pageSize=10", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, cachedData, rr.Body.String())

	mockRedis.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "FindAllPaginated")
}

func TestGetEmployeesHandlerInvalidPagination(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	// Configurar o mock do Redis para retornar um cache miss
	cacheKey := fmt.Sprintf("%s:page:%d:size:%d", employeesCacheKey, 1, 10)
	mockRedis.On("Get", mock.Anything, cacheKey).Return(newStringCmd("", redis.Nil))

	// Configurar o mock do repositório para retornar alguns dados
	mockRepo.On("FindAllPaginated", 1, 10).Return([]*domain.Employee{}, 0, nil)

	// Configurar o mock do Redis para a operação Set
	mockRedis.On("Set", mock.Anything, cacheKey, mock.Anything, 5*time.Minute).Return(newStatusCmd("OK", nil))

	handler := GetEmployeesHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("GET", "/employees?page=invalid&pageSize=invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var respBody struct {
		Employees  []*domain.Employee `json:"employees"`
		TotalCount int                `json:"totalCount"`
		Page       int                `json:"page"`
		PageSize   int                `json:"pageSize"`
		TotalPages int                `json:"totalPages"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, 1, respBody.Page)
	assert.Equal(t, 10, respBody.PageSize)

	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateEmployeeHandlerInvalidJSON(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	handler := CreateEmployeeHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("POST", "/employees", bytes.NewBufferString("invalid json"))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteEmployeeHandlerInvalidID(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	handler := DeleteEmployeeHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("DELETE", "/employees/invalid", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": "invalid",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateEmployeeHandlerInvalidID(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	handler := UpdateEmployeeHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("PUT", "/employees/invalid", bytes.NewBufferString("{}"))
	assert.NoError(t, err)

	vars := map[string]string{
		"id": "invalid",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateEmployeeHandlerInvalidJSON(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	handler := UpdateEmployeeHandler(mockRepo, mockRedis)

	req, err := http.NewRequest("PUT", "/employees/1", bytes.NewBufferString("invalid json"))
	assert.NoError(t, err)

	vars := map[string]string{
		"id": "1",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetEmployeeByIDHandlerInvalidID(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)

	handler := GetEmployeeByIDHandler(mockRepo)

	req, err := http.NewRequest("GET", "/employees/invalid", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": "invalid",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetEmployeeByIDHandlerNotFound(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)

	employeeID := 999 // An ID that doesn't exist

	mockRepo.On("FindByID", employeeID).Return(nil, nil)

	handler := GetEmployeeByIDHandler(mockRepo)

	req, err := http.NewRequest("GET", fmt.Sprintf("/employees/%d", employeeID), nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": strconv.Itoa(employeeID),
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Employee not found")

	mockRepo.AssertExpectations(t)
}

func TestCreateEmployeeHandlerRepositoryError(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	// Dados de teste
	newEmployee := &domain.Employee{
		ID:           0,
		Name:         "Alice Johnson",
		Position:     stringPtr("Engineer"),
		DepartmentID: intPtr(1),
	}

	// Simulando erro no repositório ao criar funcionário
	mockRepo.On("Create", mock.AnythingOfType("*domain.Employee")).Return(nil, fmt.Errorf("database error"))

	// Configurando o mock para invalidar o cache (não deve ser chamado)
	mockRedis.On("Del", mock.Anything, employeesCacheKey).Return(newIntCmd(0, nil)).Maybe()

	// Criando o handler
	handler := CreateEmployeeHandler(mockRepo, mockRedis)

	// Criando a requisição
	employeeJSON, _ := json.Marshal(newEmployee)
	req, err := http.NewRequest("POST", "/employees", bytes.NewBuffer(employeeJSON))
	assert.NoError(t, err)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "database error") // Verifica a mensagem de erro real

	// Verificando as expectativas dos mocks
	mockRepo.AssertExpectations(t)
	mockRedis.AssertNotCalled(t, "Del", mock.Anything, mock.Anything)
}

func TestUpdateEmployeeHandlerRepositoryError(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	employeeID := 1

	// Dados de teste
	updateData := &domain.Employee{
		Name:     "Alice Johnson Updated",
		Position: stringPtr("Senior Engineer"),
	}

	// Simulando erro no repositório ao atualizar funcionário
	mockRepo.On("Update", mock.AnythingOfType("*domain.Employee")).Return(fmt.Errorf("database update error"))

	// Configurando o mock para invalidar o cache (não deve ser chamado)
	mockRedis.On("Del", mock.Anything, employeesCacheKey).Return(newIntCmd(0, nil)).Maybe()

	// Criando o handler
	handler := UpdateEmployeeHandler(mockRepo, mockRedis)

	// Criando a requisição
	employeeJSON, _ := json.Marshal(updateData)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/employees/%d", employeeID), bytes.NewBuffer(employeeJSON))
	assert.NoError(t, err)

	// Definindo as variáveis de URL
	vars := map[string]string{
		"id": strconv.Itoa(employeeID),
	}
	req = mux.SetURLVars(req, vars)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "database update error") // Verifica a mensagem de erro real

	// Verificando as expectativas dos mocks
	mockRepo.AssertExpectations(t)
	mockRedis.AssertNotCalled(t, "Del", mock.Anything, mock.Anything)
}

func TestGetEmployeesHandlerPageOutOfRange(t *testing.T) {
	mockRepo := new(MockEmployeeRepository)
	mockRedis := new(MockRedisClient)

	// Dados de teste
	employees := []*domain.Employee{}
	totalCount := 0
	page := 5
	pageSize := 10
	totalPages := 0

	// Configurando o mock para Redis GET que retorna redis.Nil (cache miss)
	cacheKey := fmt.Sprintf("%s:page:%d:size:%d", employeesCacheKey, page, pageSize)
	mockRedis.On("Get", mock.Anything, cacheKey).Return(newStringCmd("", redis.Nil))

	// Configurando o mock para buscar do repositório
	mockRepo.On("FindAllPaginated", page, pageSize).Return(employees, totalCount, nil)

	// Configurando o mock para Redis SET
	response := struct {
		Employees  []*domain.Employee `json:"employees"`
		TotalCount int                `json:"totalCount"`
		Page       int                `json:"page"`
		PageSize   int                `json:"pageSize"`
		TotalPages int                `json:"totalPages"`
	}{
		Employees:  employees,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	jsonData, _ := json.Marshal(response)
	mockRedis.On("Set", mock.Anything, cacheKey, jsonData, 5*time.Minute).Return(newStatusCmd("OK", nil))

	// Criando o handler
	handler := GetEmployeesHandler(mockRepo, mockRedis)

	// Criando a requisição
	req, err := http.NewRequest("GET", fmt.Sprintf("/employees?page=%d&pageSize=%d", page, pageSize), nil)
	assert.NoError(t, err)

	// Gravador de resposta
	rr := httptest.NewRecorder()

	// Executando o handler
	handler.ServeHTTP(rr, req)

	// Verificando o status code (depende da implementação; pode ser OK com lista vazia ou Not Found)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"employees":[]`)

	// Verificando as expectativas dos mocks
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
