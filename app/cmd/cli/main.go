package main

import (
	cli "keeper/cmd/cli/handler"
	log "keeper/internal/logger"
	"keeper/services/keeper"
	"keeper/services/proxy"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Log struct {
		File          string        `envconfig:"LOG_FILE" default:"keeper.log"`
		Level         string        `envconfig:"LOG_LEVEL" default:"info"`
		BufSize       int           `envconfig:"LOG_BUF_SIZE" default:"100"`
		FlushInterval time.Duration `envconfig:"LOG_FLUSH_INTERVAL" default:"5s"`
	}
	Database struct {
		Name string `envconfig:"DATABASE_NAME" default:"keeper.db"`
	}
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to process env vars: %v", err)
	}

	log.Init(cfg.Log.Level, cfg.Log.File, cfg.Log.BufSize, cfg.Log.FlushInterval)

	repo, err := keeper.NewRepository(keeper.DBOptions{
		Database: cfg.Database.Name,
	})

	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}

	proxyService := proxy.New(repo)

	cli.New(repo, proxyService).Run()
}
