package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	storepkg "github.com/loaikanou/GoAuth/internal/store"
)

// PGStore is a minimal Postgres-backed store implementation for MVP.
type PGStore struct {
	DB *sql.DB
}

// Config is the Postgres connection config.
type Config struct {
	DSN string
}

// Connect opens a Postgres connection and returns a PGStore.
func Connect(cfg Config) (*PGStore, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PGStore{DB: db}, nil
}

// Close closes the underlying DB connection.
func (p *PGStore) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// CreateTenant is a placeholder implementation for Phase 2 readiness.
func (p *PGStore) CreateTenant(t *storepkg.Tenant) error {
	// Implement actual INSERT in a real implementation
	return nil
}

// GetTenant is a placeholder implementation for Phase 2 readiness.
func (p *PGStore) GetTenant(id string) (*storepkg.Tenant, error) {
	// Implement actual SELECT in a real implementation
	return &storepkg.Tenant{ID: id}, nil
}

// CreateUser is a placeholder implementation for Phase 2 readiness.
func (p *PGStore) CreateUser(u *storepkg.User) error {
	// Implement actual INSERT in a real implementation
	return nil
}

// GetUser is a placeholder implementation for Phase 2 readiness.
func (p *PGStore) GetUser(id string) (*storepkg.User, error) {
	// Implement actual SELECT in a real implementation
	return &storepkg.User{ID: id}, nil
}
