package cli

import (
	"syscall"

	_ "embed"
	"fmt"
	"os"
	"strings"

	log "keeper/internal/logger"
	"keeper/services/keeper"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

type proxyService interface {
	Start(addr string) error
	Stop() error
}

type Handler struct {
	keeper *keeper.SQLiteRepository

	proxyService proxyService
}

func New(keeper *keeper.SQLiteRepository, proxyService proxyService) *Handler {
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
				Name:   "status",
				Usage:  "Get the status of the server",
				Action: h.statusServer,
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

func (h *Handler) setKeyValue(c *cli.Context) error {
	// TODO: Implement set key-value logic
	return nil
}

func (h *Handler) getValue(c *cli.Context) error {
	log.Debugf("Getting value for key:", c.Args().First())

	return nil
}

func (h *Handler) setKeyInteractive(c *cli.Context) error {
	providerName := c.Args().First()

	fmt.Printf("Enter key for '%s': ", providerName)

	value, err := h.readSecretFromConsole()
	if err != nil {
		return log.Errorf("error reading value: %w", err)
	}

	// only show last 4 characters of the secret
	log.Debugf("Setting key %s to value %s", providerName, strings.Repeat("*", len(value)-4)+value[len(value)-4:])

	provider, err := h.keeper.GetProviderByName(c.Context, providerName)
	if err != nil {
		return log.Errorf("error getting provider: %w", err)
	}

	if _, err := h.keeper.CreateKey(c.Context, provider.Name, value); err != nil {
		return log.Errorf("error setting key: %w", err)
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
