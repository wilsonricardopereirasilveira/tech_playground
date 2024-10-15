package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pinpeople/internal/domain"
	"pinpeople/internal/repository"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// RedisClient interface
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

const employeesCacheKey = "employees"

func GetEmployeesHandler(employeeRepo repository.EmployeeRepository, rdb RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse pagination parameters
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1
		}
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 10 // Default page size
		}

		// Create a cache key that includes pagination parameters
		cacheKey := fmt.Sprintf("%s:page:%d:size:%d", employeesCacheKey, page, pageSize)

		// Try to get from cache
		cachedData, err := rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cachedData)
			return
		}

		// If not in cache, fetch from database
		employees, totalCount, err := employeeRepo.FindAllPaginated(page, pageSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a response structure that includes pagination info
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
			TotalPages: (totalCount + pageSize - 1) / pageSize,
		}

		// Serialize the data
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Store in cache for 5 minutes
		err = rdb.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Error storing in cache: %v", err)
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func CreateEmployeeHandler(employeeRepo repository.EmployeeRepository, rdb RedisClient) http.HandlerFunc {
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

		// Invalidate the cache
		ctx := r.Context()
		if err := rdb.Del(ctx, employeesCacheKey).Err(); err != nil {
			log.Printf("Erro ao invalidar o cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdEmployee)
	}
}

func DeleteEmployeeHandler(employeeRepo repository.EmployeeRepository, rdb RedisClient) http.HandlerFunc {
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

		// Invalidate the cache
		ctx := r.Context()
		if err := rdb.Del(ctx, employeesCacheKey).Err(); err != nil {
			log.Printf("Erro ao invalidar o cache: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateEmployeeHandler(employeeRepo repository.EmployeeRepository, rdb RedisClient) http.HandlerFunc {
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

		// Invalidate the cache
		ctx := r.Context()
		if err := rdb.Del(ctx, employeesCacheKey).Err(); err != nil {
			log.Printf("Erro ao invalidar o cache: %v", err)
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
