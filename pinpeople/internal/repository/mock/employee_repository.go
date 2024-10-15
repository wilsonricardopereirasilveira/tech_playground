package mock

import (
	"pinpeople/internal/domain"
)

type MockEmployeeRepository struct {
	FindAllPaginatedFunc func(page, pageSize int) ([]*domain.Employee, int, error)
	CreateFunc           func(employee *domain.Employee) (*domain.Employee, error)
	UpdateFunc           func(employee *domain.Employee) error
	DeleteFunc           func(id int) error
	FindByIDFunc         func(id int) (*domain.Employee, error)
}

func (m *MockEmployeeRepository) FindAllPaginated(page, pageSize int) ([]*domain.Employee, int, error) {
	return m.FindAllPaginatedFunc(page, pageSize)
}

func (m *MockEmployeeRepository) Create(employee *domain.Employee) (*domain.Employee, error) {
	return m.CreateFunc(employee)
}

func (m *MockEmployeeRepository) Update(employee *domain.Employee) error {
	return m.UpdateFunc(employee)
}

func (m *MockEmployeeRepository) Delete(id int) error {
	return m.DeleteFunc(id)
}

func (m *MockEmployeeRepository) FindByID(id int) (*domain.Employee, error) {
	return m.FindByIDFunc(id)
}
