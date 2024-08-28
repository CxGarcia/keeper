package keeper

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// repo.go
type Repository struct {
	db *sql.DB
}

type DBOptions struct {
	Database string
}

func NewRepository(opts DBOptions) (*Repository, error) {
	db, err := sql.Open("sqlite3", opts.Database)
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// TODO: remove
func (r *Repository) Exec(statement string) error {
	_, err := r.db.Exec(statement)

	return err
}

func (r *Repository) CreateUser(ctx context.Context, name string) (int64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", name)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetUser(ctx context.Context, id int64) (User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id)

	var user User
	if err := row.Scan(&user.Name); err != nil {
		return User{}, err
	}

	return user, nil
}

type Keys struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret"`
}

func (r *Repository) CreateKey(ctx context.Context, name, secret string) (int64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO keys (name, secret) VALUES ($1, $2)", name, secret)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) DeleteKey(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM keys WHERE id = $1", id)

	return err
}

type Provider struct {
	Keys

	ID            int64   `db:"id"`
	Name          string  `db:"name"`
	BaseURL       string  `db:"base_url"`
	Model         string  `db:"model"`
	SelectedKeyID *string `db:"selected_key_id,omitempty"`
}

func (r *Repository) CreateProviders(ctx context.Context, providers ...Provider) ([]int64, error) {
	ids := make([]int64, 0, len(providers))

	// transaction?
	for _, provider := range providers {
		res, err := r.db.ExecContext(ctx,
			`INSERT INTO providers (name, base_url, model) VALUES ($1, $2, $3)`,
			provider.Name, provider.BaseURL, provider.Model,
		)

		if err != nil {
			return nil, err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

type UpdateProviderRequest struct {
	BaseURL       string
	Model         string
	SelectedKeyID int64
}

func (r *Repository) UpdateProvider(ctx context.Context, id int64, provider UpdateProviderRequest) error {
	_, err := r.db.ExecContext(
		ctx, "UPDATE providers SET base_url = $1, selected_key_id = $2, model = $3 WHERE id = $4",
		provider.BaseURL, provider.SelectedKeyID, provider.Model, id,
	)

	return err
}

func (r *Repository) GetProvider(ctx context.Context, id int64) (Provider, error) {
	row := r.db.QueryRowContext(ctx, "SELECT name, base_url, selected_key_id, model FROM providers WHERE id = $1", id)

	var provider Provider
	if err := row.Scan(&provider.Name, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model); err != nil {
		return Provider{}, err
	}

	return provider, nil
}

func (r *Repository) GetProviderByName(ctx context.Context, name string) (Provider, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, base_url, selected_key_id, model FROM providers WHERE name = $1", name)

	var provider Provider
	if err := row.Scan(&provider.ID, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model); err != nil {
		return Provider{}, err
	}

	return provider, nil
}

func (r *Repository) GetProviderByNameWithKey(ctx context.Context, name string) (Provider, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, base_url, selected_key_id, model, keys.name, keys.secret
		FROM providers
		JOIN keys ON providers.selected_key_id = keys.id
		WHERE name = $1`, name)

	var provider Provider
	if err := row.Scan(&provider.ID, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model, &provider.Keys.Name, &provider.Secret); err != nil {
		return Provider{}, err
	}

	return provider, nil
}

func (r *Repository) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, base_url, selected_key_id, model FROM providers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	providers := make([]Provider, 0)
	for rows.Next() {
		var provider Provider
		if err := rows.Scan(&provider.Name, &provider.BaseURL, &provider.SelectedKeyID, &provider.Model); err != nil {
			return nil, err
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

type UserSettings struct {
	Provider

	UserID             int64 `db:"user_id"`
	SelectedProviderID int64 `db:"selected_provider_id"`
}

func (r *Repository) CreateUserSettings(ctx context.Context, userSettings UserSettings) (int64, error) {
	res, err := r.db.ExecContext(
		ctx, "INSERT INTO user_settings (user_id, selected_provider_id) VALUES ($1, $2)", userSettings.UserID, userSettings.SelectedProviderID,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetUserSettings(ctx context.Context, id int64) (UserSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_settings.user_id, user_settings.selected_provider_id, providers.name, providers.base_url, providers.selected_key_id, providers.model, keys.name, keys.secret
		FROM user_settings
		JOIN providers ON user_settings.selected_provider_id = providers.id
		JOIN keys ON providers.selected_key_id = keys.id
		WHERE user_settings.user_id = $1`, id)

	var userSettings UserSettings
	if err := row.Scan(
		&userSettings.UserID, &userSettings.SelectedProviderID, &userSettings.Name, &userSettings.BaseURL, &userSettings.SelectedKeyID, &userSettings.Model, &userSettings.Keys.Name, &userSettings.Secret,
	); err != nil {
		return UserSettings{}, err
	}

	return userSettings, nil
}

type UpdateUserSettingsRequest struct {
	SelectedProviderID int64
}

func (r *Repository) UpdateUserSettings(ctx context.Context, userID int64, userSettings UpdateUserSettingsRequest) error {
	_, err := r.db.ExecContext(
		ctx, "UPDATE user_settings SET selected_provider_id = $1 WHERE user_settings.user_id = $2", userSettings.SelectedProviderID, userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Close() {
	r.db.Close()
}
