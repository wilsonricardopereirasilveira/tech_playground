package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"pinpeople/internal/domain"
)

type postgresDepartmentRepository struct {
	db *sql.DB
}

func NewPostgresDepartmentRepository(db *sql.DB) *postgresDepartmentRepository {
	repo := &postgresDepartmentRepository{db}
	if err := repo.ensureTableExists(); err != nil {
		log.Fatalf("Failed to ensure departments table exists: %v", err)
	}
	return repo
}

func (r *postgresDepartmentRepository) ensureTableExists() error {
	query := `
    CREATE TABLE IF NOT EXISTS departments (
        id SERIAL PRIMARY KEY,
        company_level0 VARCHAR(255),
        company_level1 VARCHAR(255),
        company_level2 VARCHAR(255),
        company_level3 VARCHAR(255),
        company_level4 VARCHAR(255)
    );
    `
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating departments table: %v", err)
	}
	log.Println("Departments table ensured to exist")
	return nil
}

func (r *postgresDepartmentRepository) Create(department *domain.Department) (int, error) {
	query := `
        INSERT INTO departments (company_level0, company_level1, company_level2, company_level3, company_level4)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `
	var id int
	err := r.db.QueryRow(query, department.CompanyLevel0, department.CompanyLevel1, department.CompanyLevel2, department.CompanyLevel3, department.CompanyLevel4).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating department: %w", err)
	}
	return id, nil
}

func (r *postgresDepartmentRepository) FindByLevels(levels []string) (*domain.Department, error) {
	query := `
        SELECT id, company_level0, company_level1, company_level2, company_level3, company_level4
        FROM departments
        WHERE company_level0 = $1 AND company_level1 = $2 AND company_level2 = $3 AND company_level3 = $4 AND company_level4 = $5
    `
	var dept domain.Department
	err := r.db.QueryRow(query, levels[0], levels[1], levels[2], levels[3], levels[4]).Scan(
		&dept.ID, &dept.CompanyLevel0, &dept.CompanyLevel1, &dept.CompanyLevel2, &dept.CompanyLevel3, &dept.CompanyLevel4)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *postgresDepartmentRepository) FindAll() ([]*domain.Department, error) {
	query := `
        SELECT id, company_level0, company_level1, company_level2, company_level3, company_level4
        FROM departments
    `
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying departments: %w", err)
	}
	defer rows.Close()

	var departments []*domain.Department
	for rows.Next() {
		var dept domain.Department
		err := rows.Scan(&dept.ID, &dept.CompanyLevel0, &dept.CompanyLevel1, &dept.CompanyLevel2, &dept.CompanyLevel3, &dept.CompanyLevel4)
		if err != nil {
			return nil, fmt.Errorf("error scanning department row: %w", err)
		}
		departments = append(departments, &dept)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating department rows: %w", err)
	}

	return departments, nil
}
