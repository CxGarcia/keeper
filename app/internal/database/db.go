package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"keeper/internal/logger"
	provider_registry "keeper/internal/provider-registry"
	"keeper/services/keeper"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create-tables.sql
var createTablesSQL string

type Options struct {
	Database string
}

func NewSQLite(opts Options) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", opts.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func Seed(ctx context.Context, db *sql.DB, repo *keeper.SQLiteRepository, registry provider_registry.Registry) error {
	if !shouldSeed(db) {
		return nil
	}

	providers := make([]keeper.Provider, 0, len(registry.Providers))
	for _, p := range registry.Providers {
		providers = append(providers, keeper.Provider{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			Model:   p.DefaultModel,
		})
	}

	if _, err := db.Exec(createTablesSQL); err != nil {
		return logger.Errorf("failed to create tables: %v", err)
	}

	userID, err := repo.CreateProfile(ctx, keeper.CreateProfileReq{
		Name:      "default",
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		return logger.Errorf("failed to create user: %v", err)
	}

	ps, err := repo.CreateProviders(ctx, providers...)
	if err != nil {
		return logger.Errorf("failed to create providers: %v", err)
	}

	defaultProviderID := ps[0]

	if _, err := repo.CreateProfileSettings(ctx, keeper.ProfileSettings{
		ProfileID:  userID,
		ProviderID: defaultProviderID,
	}); err != nil {
		return logger.Errorf("failed to create user settings: %v", err)
	}

	return nil
}
func shouldSeed(db *sql.DB) bool {
	rows, err := db.Query(
		"SELECT name FROM sqlite_master WHERE type='table' AND name IN ('providers', 'profile_settings', 'provider_keys')",
	)
	if err != nil {
		return true
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count < 3
}
