package keeper

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// repo.go
type SqliteRepo struct {
	db *sql.DB
}

type DBOptions struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func NewRepository(opts DBOptions) (*SqliteRepo, error) {
	db, err := sql.Open("sqlite3", opts.Database)
	if err != nil {
		return nil, err
	}

	return &SqliteRepo{db: db}, nil
}

type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// TODO: remove
func (r *SqliteRepo) Exec(statement string) error {
	_, err := r.db.Exec(statement)

	return err
}

func (r *SqliteRepo) CreateUser(ctx context.Context, name string) (int64, error) {
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

func (r *SqliteRepo) GetUser(ctx context.Context, id int64) (User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id)

	var user User
	if err := row.Scan(&user.Name); err != nil {
		return User{}, err
	}

	return user, nil
}

type Provider struct {
	ID      int64   `db:"id"`
	Name    string  `db:"name"`
	BaseURL string  `db:"base_url"`
	ApiKey  *string `db:"api_key,omitempty"`
	Model   string  `db:"model"`
}

func (r *SqliteRepo) CreateProviders(ctx context.Context, providers ...Provider) ([]int64, error) {
	ids := make([]int64, 0, len(providers))

	// transaction?
	for _, provider := range providers {
		res, err := r.db.ExecContext(
			ctx, "INSERT INTO providers (name, base_url, api_key, model) VALUES ($1, $2, $3, $4)", provider.Name, provider.BaseURL, provider.ApiKey, provider.Model,
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
	baseUrl string
	apiKey  *string
	model   string
}

func (r *SqliteRepo) UpdateProvider(ctx context.Context, id int64, provider UpdateProviderRequest) error {
	_, err := r.db.ExecContext(
		ctx, "UPDATE providers SET base_url = $1, api_key = $2, model = $3 WHERE id = $4",
		provider.baseUrl, provider.apiKey, provider.model, id,
	)

	return err
}

func (r *SqliteRepo) GetProvider(ctx context.Context, id int64) (Provider, error) {
	row := r.db.QueryRowContext(ctx, "SELECT name, base_url, api_key, model FROM providers WHERE id = $1", id)

	var provider Provider
	if err := row.Scan(&provider.Name, &provider.BaseURL, &provider.ApiKey, &provider.Model); err != nil {
		return Provider{}, err
	}

	return provider, nil
}

func (r *SqliteRepo) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, base_url, api_key, model FROM providers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	providers := make([]Provider, 0)
	for rows.Next() {
		var provider Provider
		if err := rows.Scan(&provider.Name, &provider.BaseURL, &provider.ApiKey, &provider.Model); err != nil {
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

func (r *SqliteRepo) CreateUserSettings(ctx context.Context, userSettings UserSettings) (int64, error) {
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

func (r *SqliteRepo) GetUserSettings(ctx context.Context, id int64) (UserSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_settings.user_id, user_settings.selected_provider_id, providers.name, providers.base_url, providers.api_key, providers.model
		FROM user_settings
		JOIN providers ON user_settings.selected_provider_id = providers.id
		WHERE user_settings.user_id = $1`, id)

	var userSettings UserSettings
	if err := row.Scan(
		&userSettings.UserID, &userSettings.SelectedProviderID, &userSettings.Name, &userSettings.BaseURL, &userSettings.ApiKey, &userSettings.Model,
	); err != nil {
		return UserSettings{}, err
	}

	return userSettings, nil
}

type UpdateUserSettingsRequest struct {
	SelectedProviderID int64
}

func (r *SqliteRepo) UpdateUserSettings(ctx context.Context, userID int64, userSettings UpdateUserSettingsRequest) error {
	_, err := r.db.ExecContext(
		ctx, "UPDATE user_settings SET selected_provider_id = $1 WHERE user_settings.user_id = $2", userSettings.SelectedProviderID, userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *SqliteRepo) Close() {
	r.db.Close()
}

// service.go
type Service struct {
	// TODO: interface
	repo *SqliteRepo
}

func NewService(repo *SqliteRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserSettings(ctx context.Context, id int64) (UserSettings, error) {
	return s.repo.GetUserSettings(ctx, id)
}
