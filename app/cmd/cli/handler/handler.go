package cli

import (
	"fmt"
	"os"
	"strings"
	"syscall"

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
	keeper       *keeper.Service
	proxyService proxyService
}

func New(keeper *keeper.Service, proxyService proxyService) *Handler {
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
func (h *Handler) startServer(c *cli.Context) error {
	addr := fmt.Sprintf(":%s", c.String("port"))
	if err := h.proxyService.Start(addr); err != nil {
		log.Fatalf("error starting server: %v", err)
	}

	return nil
}

func (h *Handler) stopServer(c *cli.Context) error {
	if err := h.proxyService.Stop(); err != nil {
		return err
	}

	return nil
}

func (h *Handler) seedDB(c *cli.Context) error {
	// TODO: Implement seed database logic
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
	key := c.Args().First()

	fmt.Printf("Enter key for '%s': ", key)

	value, err := h.readSecretFromConsole()
	if err != nil {
		return fmt.Errorf("error reading value: %w", err)
	}

	log.Debugf("Setting key %s to value %s", key, strings.Repeat("*", len(value)))

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
