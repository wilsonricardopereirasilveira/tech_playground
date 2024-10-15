package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pinpeople/internal/domain"
	"pinpeople/internal/repository"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func GetEmployeesHandler(employeeRepo repository.EmployeeRepository, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cacheKey := "employees"

		// Tenta obter do cache
		cachedData, err := rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cachedData)
			return
		}

		// Se não estiver no cache, busca do banco de dados
		employees, err := employeeRepo.FindAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Serializa os dados
		jsonData, err := json.Marshal(employees)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Armazena no cache por 5 minutos
		err = rdb.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
		if err != nil {
			// Log do erro, mas continua a execução
			log.Printf("Erro ao armazenar no cache: %v", err)
		}

		// Envia a resposta
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func CreateEmployeeHandler(employeeRepo repository.EmployeeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var employee domain.Employee
		if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// No need to convert fields that are already pointers
		// Only convert non-pointer fields if they exist in the incoming JSON

		if employee.DepartmentID != nil {
			departmentID := *employee.DepartmentID
			employee.DepartmentID = &departmentID
		}

		// For string fields, we only need to ensure they're not empty
		if employee.Position != nil && *employee.Position != "" {
			position := *employee.Position
			employee.Position = &position
		}
		if employee.Role != nil && *employee.Role != "" {
			role := *employee.Role
			employee.Role = &role
		}
		if employee.TimeAtCompany != nil && *employee.TimeAtCompany != "" {
			timeAtCompany := *employee.TimeAtCompany
			employee.TimeAtCompany = &timeAtCompany
		}
		if employee.Gender != nil && *employee.Gender != "" {
			gender := *employee.Gender
			employee.Gender = &gender
		}
		if employee.Generation != nil && *employee.Generation != "" {
			generation := *employee.Generation
			employee.Generation = &generation
		}

		// Parse and set ResponseDate
		if employee.ResponseDate != nil {
			parsedDate, err := time.Parse("2006-01-02", employee.ResponseDate.Format("2006-01-02"))
			if err != nil {
				http.Error(w, "Invalid date format for ResponseDate", http.StatusBadRequest)
				return
			}
			employee.ResponseDate = &parsedDate
		}

		// Create the employee
		createdEmployee, err := employeeRepo.Create(&employee)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdEmployee)
	}
}

// Helper functions to convert values to pointers
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func DeleteEmployeeHandler(employeeRepo repository.EmployeeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid employee ID", http.StatusBadRequest)
			return
		}

		if err := employeeRepo.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateEmployeeHandler(employeeRepo repository.EmployeeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid employee ID", http.StatusBadRequest)
			return
		}

		var employee domain.Employee
		if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		employee.ID = id

		if err := employeeRepo.Update(&employee); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(employee)
	}
}

func GetEmployeeByIDHandler(employeeRepo repository.EmployeeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid employee ID", http.StatusBadRequest)
			return
		}

		employee, err := employeeRepo.FindByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if employee == nil {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(employee)
	}
}
