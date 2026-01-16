package database

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    picture TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
