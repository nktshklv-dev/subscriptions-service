package data

import (
	"database/sql"
	"time"
)

func NewDB(connString string, maxOpen, maxIdle int, maxLife time.Duration) (*sql.DB, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}
	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}
	if maxLife > 0 {
		db.SetConnMaxLifetime(maxLife)
	}
	return db, nil
}
