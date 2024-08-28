package cli

import (
	"syscall"

	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	log "keeper/internal/logger"
	"keeper/services/keeper"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type proxyService interface {
	Start(addr string) error
	Stop() error
}

type Handler struct {
	keeper *keeper.Repository

	proxyService proxyService
}

func New(keeper *keeper.Repository, proxyService proxyService) *Handler {
	return &Handler{
		keeper:       keeper,
		proxyService: proxyService,
	}
}

func (h *Handler) Run() error {
	app := &cli.App{
		Name:  "keeper",
		Usage: "A CLI tool for managing key-value pairs",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   "8080",
						Usage:   "Port to run the server on",
					},
					&cli.BoolFlag{
						Name:    "detached",
						Aliases: []string{"d"},
						Value:   false,
						Usage:   "Run the server in detached mode",
					},
				},
				Action: h.startServer,
			},
			{
				Name:   "stop",
				Usage:  "Stop the server",
				Action: h.stopServer,
			},
			{
				Name:   "seed-db",
				Usage:  "Seed the database",
				Action: h.seedDB,
			},
			{
				Name:      "set",
				Usage:     "Set a key-value pair",
				ArgsUsage: "<key> <value>",
				Action:    h.setKeyValue,
			},
			{
				Name:      "get",
				Usage:     "Get a value by key",
				ArgsUsage: "<key>",
				Action:    h.getValue,
			},
			{
				Name:   "set-key",
				Usage:  "Set a key-value pair interactively",
				Action: h.setKeyInteractive,
			},
		},
	}

	return app.Run(os.Args)
}

func (h *Handler) seedDB(c *cli.Context) error {
	seedDBBBB(h.keeper)

	return nil
}

func (h *Handler) setKeyValue(c *cli.Context) error {
	// TODO: Implement set key-value logic
	return nil
}

func (h *Handler) getValue(c *cli.Context) error {
	log.Debugf("Getting value for key:", c.Args().First())

	return nil
}

func (h *Handler) setKeyInteractive(c *cli.Context) error {
	provider := c.Args().First()

	fmt.Printf("Enter key for '%s': ", provider)

	value, err := h.readSecretFromConsole()
	if err != nil {
		return fmt.Errorf("error reading value: %w", err)
	}

	log.Debugf("Setting key %s to value %s", provider, strings.Repeat("*", len(value)))

	if _, err := h.keeper.CreateKey(c.Context, "xxx", value); err != nil {
		return fmt.Errorf("error setting key: %w", err)
	}

	return nil
}

func (h *Handler) readSecretFromConsole() (string, error) {
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()

	return string(password), nil
}

//go:embed provider-registry.yml
var registry []byte

//go:embed create-tables.sql
var createTablesSQL string

func seedDBBBB(repo *keeper.Repository) {
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
