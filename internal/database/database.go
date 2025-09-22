package database

import (
	"fmt"
	"subscription-aggregator/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Retry logic for database connection
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to ping database after %d retries: %w", maxRetries, err)
			}
			time.Sleep(retryInterval)
			continue
		}
		break
	}

	return db, nil
}
