package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"pinpeople/internal/domain"
)

type postgresLocationRepository struct {
	db *sql.DB
}

func NewPostgresLocationRepository(db *sql.DB) *postgresLocationRepository {
	repo := &postgresLocationRepository{db}
	if err := repo.ensureTableExists(); err != nil {
		log.Fatalf("Failed to ensure locations table exists: %v", err)
	}
	return repo
}

func (r *postgresLocationRepository) ensureTableExists() error {
	query := `
    CREATE TABLE IF NOT EXISTS locations (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );
    `
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating locations table: %v", err)
	}
	log.Println("Locations table ensured to exist")
	return nil
}

func (r *postgresLocationRepository) Create(location *domain.Location) (int, error) {
	query := `INSERT INTO locations (name) VALUES ($1) RETURNING id`
	var id int
	err := r.db.QueryRow(query, location.Name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating location: %v", err)
	}
	return id, nil
}

func (r *postgresLocationRepository) FindByName(name string) (*domain.Location, error) {
	query := `SELECT id, name FROM locations WHERE name = $1`
	var location domain.Location
	err := r.db.QueryRow(query, name).Scan(&location.ID, &location.Name)
	if err != nil {
		return nil, err
	}
	return &location, nil
}
