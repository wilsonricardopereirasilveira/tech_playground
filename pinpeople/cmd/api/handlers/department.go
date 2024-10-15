package handlers

import (
	"encoding/json"
	"net/http"
	"pinpeople/internal/domain"
	"pinpeople/internal/repository"
)

func GetDepartmentsHandler(departmentRepo repository.DepartmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		departments, err := departmentRepo.FindAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(departments)
	}
}

func CreateDepartmentHandler(departmentRepo repository.DepartmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var department domain.Department
		if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		id, err := departmentRepo.Create(&department)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		department.ID = id
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(department)
	}
}
