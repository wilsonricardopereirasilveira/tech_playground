package main

import (
    "database/sql"
    "log"
    "os"
    "pinpeople/internal/repository/postgres"
    "pinpeople/internal/usecase"

    _ "github.com/lib/pq"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=db user=pinpeople_user password=PinP_s3cur3_p@ssw0rd dbname=pinpeople_db sslmode=disable"
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Criar repositórios na ordem correta
    locationRepo := postgres.NewPostgresLocationRepository(db) // Certifique-se de que a tabela locations existe
    departmentRepo := postgres.NewPostgresDepartmentRepository(db) // Certifique-se de que a tabela departments existe
    employeeRepo := postgres.NewPostgresEmployeeRepository(db) // Isso deve ocorrer após as tabelas anteriores

    useCase := usecase.NewImportEmployeesUseCase(employeeRepo, departmentRepo, locationRepo)

    err = useCase.Execute("data/employees.csv")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Import completed successfully")
}
