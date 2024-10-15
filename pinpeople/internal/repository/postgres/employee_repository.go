package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"pinpeople/internal/domain"
)

type postgresEmployeeRepository struct {
	db *sql.DB
}

func NewPostgresEmployeeRepository(db *sql.DB) *postgresEmployeeRepository {
	repo := &postgresEmployeeRepository{db}
	if err := repo.ensureTableExists(); err != nil {
		log.Fatalf("Failed to ensure employees table exists: %v", err)
	}
	return repo
}

func (r *postgresEmployeeRepository) ensureTableExists() error {
	query := `
    CREATE TABLE IF NOT EXISTS employees (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255),
        email VARCHAR(255),
        corporate_email VARCHAR(255),
        department_id INT REFERENCES departments(id),
        position VARCHAR(100),
        role VARCHAR(100),
        location_id INT REFERENCES locations(id),
        time_at_company VARCHAR(50),
        gender VARCHAR(50),
        generation VARCHAR(50),
        response_date DATE,
        position_interest INT,
        position_interest_comments TEXT,
        contribution INT,
        contribution_comments TEXT,
        learning_development INT,
        learning_development_comments TEXT,
        feedback INT,
        feedback_comments TEXT,
        manager_interaction INT,
        manager_interaction_comments TEXT,
        career_clarity INT,
        career_clarity_comments TEXT,
        retention_expectation INT,
        retention_expectation_comments TEXT,
        enps INT,
        enps_comments TEXT,
        open_enps TEXT
    );
    `
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating employees table: %v", err)
	}
	log.Println("Employees table ensured to exist")
	return nil
}

func (r *postgresEmployeeRepository) Create(employee *domain.Employee) (*domain.Employee, error) {
	query := `
        INSERT INTO employees (
            name, email, corporate_email, department_id, position, role, location_id,
            time_at_company, gender, generation, response_date, position_interest,
            position_interest_comments, contribution, contribution_comments,
            learning_development, learning_development_comments, feedback, feedback_comments,
            manager_interaction, manager_interaction_comments, career_clarity,
            career_clarity_comments, retention_expectation, retention_expectation_comments,
            enps, enps_comments, open_enps
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
            $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
        ) RETURNING id`

	err := r.db.QueryRow(query,
		employee.Name, employee.Email, employee.CorporateEmail, employee.DepartmentID,
		employee.Position, employee.Role, employee.LocationID, employee.TimeAtCompany,
		employee.Gender, employee.Generation, employee.ResponseDate, employee.PositionInterest,
		employee.PositionInterestComments, employee.Contribution, employee.ContributionComments,
		employee.LearningDevelopment, employee.LearningDevelopmentComments, employee.Feedback,
		employee.FeedbackComments, employee.ManagerInteraction, employee.ManagerInteractionComments,
		employee.CareerClarity, employee.CareerClarityComments, employee.RetentionExpectation,
		employee.RetentionExpectationComments, employee.ENPS, employee.ENPSComments, employee.OpenENPS).Scan(&employee.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating employee: %w", err)
	}
	return employee, nil
}

func (r *postgresEmployeeRepository) FindAll() ([]*domain.Employee, error) {
	query := `SELECT * FROM employees`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		var employee domain.Employee
		err := rows.Scan(
			&employee.ID, &employee.Name, &employee.Email, &employee.CorporateEmail, &employee.DepartmentID,
			&employee.Position, &employee.Role, &employee.LocationID, &employee.TimeAtCompany, &employee.Gender,
			&employee.Generation, &employee.ResponseDate, &employee.PositionInterest, &employee.PositionInterestComments,
			&employee.Contribution, &employee.ContributionComments, &employee.LearningDevelopment,
			&employee.LearningDevelopmentComments, &employee.Feedback, &employee.FeedbackComments,
			&employee.ManagerInteraction, &employee.ManagerInteractionComments, &employee.CareerClarity,
			&employee.CareerClarityComments, &employee.RetentionExpectation, &employee.RetentionExpectationComments,
			&employee.ENPS, &employee.ENPSComments, &employee.OpenENPS,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, &employee)
	}

	return employees, nil
}

func (r *postgresEmployeeRepository) FindByID(id int) (*domain.Employee, error) {
	var employee domain.Employee
	query := `SELECT * FROM employees WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&employee.ID, &employee.Name, &employee.Email, &employee.CorporateEmail, &employee.DepartmentID,
		&employee.Position, &employee.Role, &employee.LocationID, &employee.TimeAtCompany, &employee.Gender,
		&employee.Generation, &employee.ResponseDate, &employee.PositionInterest, &employee.PositionInterestComments,
		&employee.Contribution, &employee.ContributionComments, &employee.LearningDevelopment,
		&employee.LearningDevelopmentComments, &employee.Feedback, &employee.FeedbackComments,
		&employee.ManagerInteraction, &employee.ManagerInteractionComments, &employee.CareerClarity,
		&employee.CareerClarityComments, &employee.RetentionExpectation, &employee.RetentionExpectationComments,
		&employee.ENPS, &employee.ENPSComments, &employee.OpenENPS,
	)
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *postgresEmployeeRepository) Update(employee *domain.Employee) error {
	query := `
        UPDATE employees SET
            name = $1, email = $2, corporate_email = $3, department_id = $4, position = $5,
            role = $6, location_id = $7, time_at_company = $8, gender = $9, generation = $10,
            response_date = $11, position_interest = $12, position_interest_comments = $13,
            contribution = $14, contribution_comments = $15, learning_development = $16,
            learning_development_comments = $17, feedback = $18, feedback_comments = $19,
            manager_interaction = $20, manager_interaction_comments = $21, career_clarity = $22,
            career_clarity_comments = $23, retention_expectation = $24, retention_expectation_comments = $25,
            enps = $26, enps_comments = $27, open_enps = $28
        WHERE id = $29
    `
	_, err := r.db.Exec(query,
		employee.Name, employee.Email, employee.CorporateEmail, employee.DepartmentID, employee.Position,
		employee.Role, employee.LocationID, employee.TimeAtCompany, employee.Gender, employee.Generation,
		employee.ResponseDate, employee.PositionInterest, employee.PositionInterestComments, employee.Contribution,
		employee.ContributionComments, employee.LearningDevelopment, employee.LearningDevelopmentComments,
		employee.Feedback, employee.FeedbackComments, employee.ManagerInteraction,
		employee.ManagerInteractionComments, employee.CareerClarity, employee.CareerClarityComments,
		employee.RetentionExpectation, employee.RetentionExpectationComments, employee.ENPS,
		employee.ENPSComments, employee.OpenENPS, employee.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating employee: %w", err)
	}
	return nil
}

func (r *postgresEmployeeRepository) Delete(id int) error {
	query := `DELETE FROM employees WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting employee: %w", err)
	}
	return nil
}
