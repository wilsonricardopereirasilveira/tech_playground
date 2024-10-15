package repository

import "pinpeople/internal/domain"

type EmployeeRepository interface {
	Create(employee *domain.Employee) (*domain.Employee, error)
	FindAll() ([]*domain.Employee, error)
	FindByID(id int) (*domain.Employee, error)
	Update(employee *domain.Employee) error
	Delete(id int) error
}

type DepartmentRepository interface {
	Create(department *domain.Department) (int, error)
	FindAll() ([]*domain.Department, error)
	FindByLevels(levels []string) (*domain.Department, error)
}

type LocationRepository interface {
	Create(location *domain.Location) (int, error)
	FindByName(name string) (*domain.Location, error)
}
