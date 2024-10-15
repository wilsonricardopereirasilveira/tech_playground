package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"pinpeople/cmd/api/handlers"
	"pinpeople/internal/middleware"
	"pinpeople/internal/repository/postgres"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db := setupDatabase()
	defer db.Close()

	rdb := setupRedis()
	defer rdb.Close()

	router := setupRouter(db, rdb)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func setupDatabase() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func setupRouter(db *sql.DB, rdb *redis.Client) *mux.Router {
	router := mux.NewRouter()

	employeeRepo := postgres.NewPostgresEmployeeRepository(db)
	departmentRepo := postgres.NewPostgresDepartmentRepository(db)

	// Rotas públicas
	router.HandleFunc("/ping", handlers.PingHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Rotas protegidas
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JWTMiddleware)
	api.HandleFunc("/employees", handlers.GetEmployeesHandler(employeeRepo, rdb)).Methods("GET")
	api.HandleFunc("/employees", handlers.CreateEmployeeHandler(employeeRepo)).Methods("POST")
	api.HandleFunc("/employees/{id:[0-9]+}", handlers.UpdateEmployeeHandler(employeeRepo)).Methods("PUT")
	api.HandleFunc("/employees/{id:[0-9]+}", handlers.DeleteEmployeeHandler(employeeRepo)).Methods("DELETE")

	api.HandleFunc("/departments", handlers.GetDepartmentsHandler(departmentRepo)).Methods("GET")
	api.HandleFunc("/departments", handlers.CreateDepartmentHandler(departmentRepo)).Methods("POST")

	return router
}
