package main

import (
	"context"
	_ "embed"
	"fmt"
	"keeper/cmd/api/httpjson"
	"keeper/services/keeper"

	"gopkg.in/yaml.v3"
)

func main() {
	repo, err := keeper.NewRepository(keeper.DBOptions{
		Host:     "localhost",
		Port:     5432,
		User:     "sqlite",
		Password: "password",
		Database: "example",
	})
	if err != nil {
		panic(err)
	}

	keeper := keeper.NewService(repo)

	h := httpjson.New(keeper)
	h.Start(":8080")
}

//go:embed provider-registry.yml
var registry []byte

//go:embed create-tables.sql
var createTablesSQL string

func seedDB(repo *keeper.SqliteRepo) {
	ctx := context.Background()

	registry, err := loadRegistry()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	providers := make([]keeper.Provider, 0, len(registry.Providers))
	for _, p := range registry.Providers {
		providers = append(providers, keeper.Provider{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			Model:   p.DefaultModel,
		})
	}

	fmt.Print(providers)

	fmt.Println(createTablesSQL)

	if err := repo.Exec(createTablesSQL); err != nil {
		fmt.Println(err)
		panic(err)
	}

	userID, err := repo.CreateUser(ctx, "Keeper")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	ps, err := repo.CreateProviders(ctx, providers...)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defaultProviderID := ps[0]

	if _, err := repo.CreateUserSettings(ctx, keeper.UserSettings{
		UserID:             userID,
		SelectedProviderID: defaultProviderID,
	}); err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type Model struct {
	Name string `json:"name"`
}

type ProviderAuth struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Provider struct {
	Name         string       `json:"name"`
	BaseURL      string       `json:"base_url"`
	DefaultModel string       `json:"default_model"`
	Models       []Model      `json:"models"`
	Auth         ProviderAuth `json:"auth"`
}

type ProviderRegistry struct {
	Providers []Provider `json:"providers"`
}

func loadRegistry() (ProviderRegistry, error) {
	var reg ProviderRegistry
	if err := yaml.Unmarshal(registry, &reg); err != nil {
		return ProviderRegistry{}, err
	}

	return reg, nil
}
