package usecase

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"pinpeople/internal/domain"
	"pinpeople/internal/repository"
	"strconv"
	"time"
)

type ImportEmployeesUseCase struct {
	employeeRepo repository.EmployeeRepository
	deptRepo     repository.DepartmentRepository
	locationRepo repository.LocationRepository
}

func NewImportEmployeesUseCase(
	employeeRepo repository.EmployeeRepository,
	deptRepo repository.DepartmentRepository,
	locationRepo repository.LocationRepository,
) *ImportEmployeesUseCase {
	return &ImportEmployeesUseCase{employeeRepo, deptRepo, locationRepo}
}

func (uc *ImportEmployeesUseCase) Execute(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	// Skip the header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}

		departmentID, err := uc.getOrCreateDepartment(record[10:15])
		if err != nil {
			return fmt.Errorf("error creating department: %w", err)
		}

		locationID, err := uc.getOrCreateLocation(record[6])
		if err != nil {
			return fmt.Errorf("error creating location: %w", err)
		}

		employee, err := createEmployeeFromRecord(record, departmentID, locationID)
		if err != nil {
			return fmt.Errorf("error creating employee from record: %w", err)
		}

		createdEmployee, err := uc.employeeRepo.Create(employee)
		if err != nil {
			return fmt.Errorf("error creating employee in repository: %w", err)
		}

		// You can log or do something with the createdEmployee if needed
		_ = createdEmployee
	}

	return nil
}

func createEmployeeFromRecord(record []string, departmentID int, locationID int) (*domain.Employee, error) {
	employee := &domain.Employee{
		Name:           record[0],
		Email:          record[1],
		CorporateEmail: record[2],
		DepartmentID:   &departmentID,
		Position:       parseNullableString(record[4]),
		Role:           parseNullableString(record[5]),
		LocationID:     &locationID,
		TimeAtCompany:  parseNullableString(record[7]),
		Gender:         parseNullableString(record[8]),
		Generation:     parseNullableString(record[9]),
	}

	if record[15] != "" {
		responseDate, err := time.Parse("02/01/2006", record[15])
		if err != nil {
			return nil, fmt.Errorf("invalid date format for ResponseDate: %w", err)
		}
		employee.ResponseDate = &responseDate
	}

	employee.PositionInterest = parseNullableInt(record[16])
	employee.PositionInterestComments = parseNullableString(record[17])

	employee.Contribution = parseNullableInt(record[18])
	employee.ContributionComments = parseNullableString(record[19])

	employee.LearningDevelopment = parseNullableInt(record[20])
	employee.LearningDevelopmentComments = parseNullableString(record[21])

	employee.Feedback = parseNullableInt(record[22])
	employee.FeedbackComments = parseNullableString(record[23])

	employee.ManagerInteraction = parseNullableInt(record[24])
	employee.ManagerInteractionComments = parseNullableString(record[25])

	employee.CareerClarity = parseNullableInt(record[26])
	employee.CareerClarityComments = parseNullableString(record[27])

	employee.RetentionExpectation = parseNullableInt(record[28])
	employee.RetentionExpectationComments = parseNullableString(record[29])

	employee.ENPS = parseNullableInt(record[30])
	employee.ENPSComments = parseNullableString(record[31])

	if len(record) > 32 {
		employee.OpenENPS = parseNullableString(record[32])
	}

	return employee, nil
}

func parseNullableInt(s string) *int {
	if s == "" {
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

func parseNullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (uc *ImportEmployeesUseCase) getOrCreateDepartment(deptData []string) (int, error) {
	if len(deptData) != 5 {
		return 0, fmt.Errorf("invalid department data: expected 5 levels, got %d", len(deptData))
	}

	dept, err := uc.deptRepo.FindByLevels(deptData)
	if err == sql.ErrNoRows {
		newDept := &domain.Department{
			CompanyLevel0: deptData[0],
			CompanyLevel1: deptData[1],
			CompanyLevel2: deptData[2],
			CompanyLevel3: deptData[3],
			CompanyLevel4: deptData[4],
		}
		deptID, err := uc.deptRepo.Create(newDept)
		if err != nil {
			return 0, fmt.Errorf("error creating new department: %w", err)
		}
		return deptID, nil
	} else if err != nil {
		return 0, fmt.Errorf("error finding department: %w", err)
	}

	if dept == nil {
		return 0, fmt.Errorf("unexpected nil department returned")
	}

	return dept.ID, nil
}

func (uc *ImportEmployeesUseCase) getOrCreateLocation(location string) (int, error) {
	loc, err := uc.locationRepo.FindByName(location)
	if err != nil {
		if err == sql.ErrNoRows {
			newLocation := &domain.Location{Name: location}
			return uc.locationRepo.Create(newLocation)
		}
		return 0, fmt.Errorf("error finding location: %w", err)
	}

	if loc == nil {
		return 0, fmt.Errorf("unexpected nil location for '%s'", location)
	}

	return loc.ID, nil
}
