package migrate

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

const (
	dialectPostgres = "postgres"
)

func Up(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect(dialectPostgres); err != nil {
		return err
	}
	return goose.Up(db, migrationsDir)
}
