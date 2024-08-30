package keeper

import (
	"context"
	"database/sql"
	"fmt"
	"keeper/internal/logger"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

type Profile struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	IsActive  bool   `db:"is_active"`
	IsDefault bool   `db:"is_default"`
}

// ProviderKey
type ProviderKey struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret"`
}

// Provider
type Provider struct {
	ProviderKey

	ID            int64   `db:"id"`
	Name          string  `db:"name"`
	BaseURL       string  `db:"base_url"`
	Model         string  `db:"model"`
	SelectedKeyID *string `db:"selected_key_id,omitempty"`
}

type UpdateProviderRequest struct {
	BaseURL       string
	Model         string
	SelectedKeyID int64
}

type ProfileSettings struct {
	Provider

	ProfileID  int64 `db:"user_id"`
	ProviderID int64 `db:"selected_provider_id"`
}

type UpdateUserSettingsRequest struct {
	SelectedProviderID int64
}

func NewSQLite(db *sql.DB) (*SQLiteRepository, error) {
	return &SQLiteRepository{db: db}, nil
}

type CreateProfileReq struct {
	Name      string
	IsActive  bool
	IsDefault bool
}

func (r *SQLiteRepository) CreateProfile(ctx context.Context, req CreateProfileReq) (int64, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO profiles (name, is_active, is_default) VALUES ($1, $2, $3)", req.Name, req.IsActive, req.IsDefault)
	if err != nil {
		return 0, logger.Errorf("failed to create profile: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// Key repository
func (r *SQLiteRepository) CreateProviderKey(ctx context.Context, provider Provider, secret string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, logger.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the new provider key
	result, err := tx.ExecContext(ctx, "INSERT INTO provider_keys (provider_id, name, secret) VALUES ($1, $2, $3)", provider.ID, provider.Name, secret)
	if err != nil {
		return 0, logger.Errorf("failed to create key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	// Check if the active profile has a provider key for this provider
	var count int
	err = tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM profile_settings ps
        JOIN profiles p ON ps.profile_id = p.id
        WHERE p.is_active = 1 AND ps.provider_id = $1 AND ps.provider_key_id IS NOT NULL
    `, provider.ID).Scan(&count)
	if err != nil {
		return 0, logger.Errorf("failed to check existing provider key: %w", err)
	}

	// If no provider key is set for the active profile, set this new key
	if count == 0 {
		_, err = tx.ExecContext(ctx, `
            UPDATE profile_settings
            SET provider_key_id = $1
            WHERE profile_id = (SELECT id FROM profiles WHERE is_active = 1)
            AND provider_id = $2
        `, id, provider.ID)
		if err != nil {
			return 0, logger.Errorf("failed to update profile settings with new key: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, logger.Errorf("failed to commit transaction: %w", err)
	}

	return id, nil
}

// Provider repository
func (r *SQLiteRepository) CreateProviders(ctx context.Context, providers ...Provider) ([]int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, logger.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	ids := make([]int64, 0, len(providers))
	for _, provider := range providers {
		if provider.Name == "" || provider.BaseURL == "" || provider.Model == "" {
			return nil, logger.Errorf("provider name, base URL, and model cannot be empty")
		}

		result, err := tx.ExecContext(ctx, "INSERT INTO providers (name, base_url, model) VALUES ($1, $2, $3)", provider.Name, provider.BaseURL, provider.Model)
		if err != nil {
			return nil, logger.Errorf("failed to create provider: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, logger.Errorf("failed to get last insert ID: %w", err)
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, logger.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

func (r *SQLiteRepository) GetProviderByName(ctx context.Context, name string) (*Provider, error) {
	if name == "" {
		return nil, logger.Errorf("provider name cannot be empty")
	}

	var provider Provider
	err := r.db.QueryRowContext(ctx, "SELECT id, base_url, model FROM providers WHERE name = $1", name).Scan(&provider.ID, &provider.BaseURL, &provider.Model)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, logger.Errorf("provider not found")
		default:
			return nil, logger.Errorf("failed to get provider: %w", err)
		}
	}

	provider.Name = name
	return &provider, nil
}

func (r *SQLiteRepository) GetProviderByNameWithKey(ctx context.Context, name string) (*Provider, error) {
	if name == "" {
		return nil, logger.Errorf("provider name cannot be empty")
	}

	var provider Provider
	var keyID sql.NullInt64
	var keyName, secret sql.NullString

	err := r.db.QueryRowContext(ctx, `
        SELECT p.id, p.name, p.base_url, p.model,
               k.id, k.name, k.secret
        FROM providers p
        LEFT JOIN provider_keys k ON k.provider_id = p.id AND k.is_active = 1
        WHERE p.name = $1
        ORDER BY k.id DESC
        LIMIT 1`, name).Scan(
		&provider.ID, &provider.Name, &provider.BaseURL, &provider.Model,
		&keyID, &keyName, &secret,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, logger.Errorf("provider not found")
		default:
			return nil, logger.Errorf("failed to get provider: %w", err)
		}
	}

	if keyID.Valid {
		provider.SelectedKeyID = new(string)
		*provider.SelectedKeyID = fmt.Sprintf("%d", keyID.Int64)
		provider.ProviderKey = ProviderKey{
			ID:     keyID.Int64,
			Name:   keyName.String,
			Secret: secret.String,
		}
	}

	return &provider, nil
}

// User settings repository
func (r *SQLiteRepository) CreateProfileSettings(ctx context.Context, userSettings ProfileSettings) (int64, error) {
	if userSettings.ProfileID <= 0 || userSettings.ProviderID <= 0 {
		return 0, logger.Errorf("invalid user ID or selected provider ID")
	}

	result, err := r.db.ExecContext(ctx, "INSERT INTO profile_settings (profile_id, provider_id) VALUES ($1, $2)", userSettings.ProfileID, userSettings.ProviderID)
	if err != nil {
		return 0, logger.Errorf("failed to create user settings: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *SQLiteRepository) GetActiveProfileSettingsWithKey(ctx context.Context) (*ProfileSettings, error) {

	var settings ProfileSettings
	var providerKeyID, providerID sql.NullInt64
	var providerName, providerBaseURL, providerModel, keyName, keySecret sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT ps.profile_id, ps.provider_id, ps.provider_key_id, p.name, p.base_url, p.model, k.name, k.secret
		FROM profile_settings ps
		LEFT JOIN providers p ON ps.provider_id = p.id
		LEFT JOIN provider_keys k ON ps.provider_key_id = k.id
		WHERE ps.profile_id = (
			SELECT id
			FROM profiles
			WHERE is_active = 1
			LIMIT 1
		)`).
		Scan(
			&settings.ProfileID, &providerID, &providerKeyID,
			&providerName, &providerBaseURL, &providerModel,
			&keyName, &keySecret,
		)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, logger.Errorf("profile settings not found")

		default:
			return nil, logger.Errorf("failed to get profile settings: %w", err)
		}
	}

	// Set Provider details if available
	if providerID.Valid {
		settings.ProviderID = providerID.Int64
		settings.Provider = Provider{
			ID:      providerID.Int64,
			Name:    providerName.String,
			BaseURL: providerBaseURL.String,
			Model:   providerModel.String,
		}
	}

	// Set ProviderKey details if available
	if providerKeyID.Valid {
		settings.Provider.ProviderKey = ProviderKey{
			ID:     providerKeyID.Int64,
			Name:   keyName.String,
			Secret: keySecret.String,
		}
	}

	return &settings, nil
}
