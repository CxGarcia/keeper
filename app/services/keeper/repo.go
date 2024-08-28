package keeper

import (
	"context"
	"database/sql"
	"keeper/internal/logger"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

// User
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// Key
type Key struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret"`
}

// Provider
type Provider struct {
	Key

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

// UserSettings
type UserSettings struct {
	Provider

	UserID             int64 `db:"user_id"`
	SelectedProviderID int64 `db:"selected_provider_id"`
}

type UpdateUserSettingsRequest struct {
	SelectedProviderID int64
}

func NewSQLite(db *sql.DB) (*SQLiteRepository, error) {
	return &SQLiteRepository{db: db}, nil
}

// User repository
func (r *SQLiteRepository) CreateUser(ctx context.Context, name string) (int64, error) {
	if name == "" {
		return 0, logger.Errorf("name cannot be empty")
	}

	result, err := r.db.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", name)
	if err != nil {
		return 0, logger.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *SQLiteRepository) GetUser(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, logger.Errorf("invalid user ID")
	}

	var user User
	err := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id).Scan(&user.Name)
	if err != nil {
		switch {
			case err == sql.ErrNoRows:
				return nil, logger.Errorf("user not found")
			default:
				return nil, logger.Errorf("failed to get user: %w", err
		}
	}

	user.ID = id
	return &user, nil
}

// Key repository
func (r *SQLiteRepository) CreateKey(ctx context.Context, name, secret string) (int64, error) {
	if secret == "" {
		return 0, logger.Errorf("secret cannot be empty")
	}

	result, err := r.db.ExecContext(ctx, "INSERT INTO provider_keys (name, secret) VALUES ($1, $2)", name, secret)
	if err != nil {
		return 0, logger.Errorf("failed to create key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *SQLiteRepository) DeleteKey(ctx context.Context, id int64) error {
	if id <= 0 {
		return logger.Errorf("invalid key ID")
	}

	_, err := r.db.ExecContext(ctx, "DELETE FROM provider_keys WHERE id = $1", id)
	if err != nil {
		return logger.Errorf("failed to delete key: %w", err)
	}

	return nil
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

func (r *SQLiteRepository) UpdateProvider(ctx context.Context, id int64, provider UpdateProviderRequest) error {
	if id <= 0 {
		return logger.Errorf("invalid provider ID")
	}

	_, err := r.db.ExecContext(ctx, "UPDATE providers SET base_url = $1, selected_key_id = $2, model = $3 WHERE id = $4", provider.BaseURL, provider.SelectedKeyID, provider.Model, id)
	if err != nil {
		return logger.Errorf("failed to update provider: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetProvider(ctx context.Context, id int64) (*Provider, error) {
	if id <= 0 {
		return nil, logger.Errorf("invalid provider ID")
	}

	var provider Provider
	err := r.db.QueryRowContext(ctx, "SELECT name, base_url, selected_key_id, model FROM providers WHERE id = $1", id).Scan(&provider.Name, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, logger.Errorf("provider not found")
		default:
			return nil, logger.Errorf("failed to get provider: %w", err)
		}
	}

	provider.ID = id
	return &provider, nil
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
	err := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.base_url, p.selected_key_id, p.model, k.name, k.secret
		FROM providers p
		JOIN provider_keys k ON p.selected_key_id = k.id
		WHERE p.name = $1`, name).Scan(
		&provider.ID, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model, &provider.Key.Name, &provider.Secret,
	)
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

func (r *SQLiteRepository) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, base_url, selected_key_id, model FROM providers")
	if err != nil {
		return nil, logger.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		var provider Provider
		if err := rows.Scan(&provider.Name, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model); err != nil {
			return nil, logger.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, logger.Errorf("error iterating providers: %w", err)
	}

	return providers, nil
}

// User settings repository
func (r *SQLiteRepository) CreateUserSettings(ctx context.Context, userSettings UserSettings) (int64, error) {
	if userSettings.UserID <= 0 || userSettings.SelectedProviderID <= 0 {
		return 0, logger.Errorf("invalid user ID or selected provider ID")
	}

	result, err := r.db.ExecContext(ctx, "INSERT INTO user_settings (user_id, selected_provider_id) VALUES ($1, $2)", userSettings.UserID, userSettings.SelectedProviderID)
	if err != nil {
		return 0, logger.Errorf("failed to create user settings: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, logger.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *SQLiteRepository) GetUserSettings(ctx context.Context, id int64) (*UserSettings, error) {
	if id <= 0 {
		return nil, logger.Errorf("invalid user ID")
	}

	var userSettings UserSettings
	err := r.db.QueryRowContext(ctx, `
		SELECT us.user_id, us.selected_provider_id, p.name, p.base_url, p.selected_key_id, p.model, k.name, k.secret
		FROM user_settings us
		JOIN providers p ON us.selected_provider_id = p.id
		JOIN provider_keys k ON p.selected_key_id = k.id
		WHERE us.user_id = $1`, id).Scan(
		&userSettings.UserID, &userSettings.SelectedProviderID, &userSettings.Name, &userSettings.BaseURL,
		&userSettings.SelectedKeyID, &userSettings.Model, &userSettings.Key.Name, &userSettings.Secret,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, logger.Errorf("user settings not found")
		default:
			return nil, logger.Errorf("failed to get user settings: %w", err)
		}
	}

	return &userSettings, nil
}

func (r *SQLiteRepository) UpdateUserSettings(ctx context.Context, userID int64, userSettings UpdateUserSettingsRequest) error {
	if userID <= 0 || userSettings.SelectedProviderID <= 0 {
		return logger.Errorf("invalid user ID or selected provider ID")
	}

	_, err := r.db.ExecContext(ctx, "UPDATE user_settings SET selected_provider_id = $1 WHERE user_id = $2", userSettings.SelectedProviderID, userID)
	if err != nil {
		return logger.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
