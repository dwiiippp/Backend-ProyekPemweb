package config

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql" // driver MySQL
)

// OpenDB membuka pool koneksi MySQL dan memastikan database bisa dijangkau.
func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
