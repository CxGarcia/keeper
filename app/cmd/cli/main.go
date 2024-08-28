package main

import (
	"context"
	cli "keeper/cmd/cli/handler"
	"keeper/internal/database"
	log "keeper/internal/logger"
	provider_registry "keeper/internal/provider-registry"
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
	ctx := context.Background()

	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to process env vars: %v", err)
	}

	log.Init(
		cfg.Log.Level, cfg.Log.File, cfg.Log.BufSize, cfg.Log.FlushInterval,
	)

	defer log.Close()

	reg, err := provider_registry.New()
	if err != nil {
		log.Fatalf("failed to load provider registry: %v", err)
	}

	db, err := database.NewSQLite(database.Options{Database: cfg.Database.Name})
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	defer db.Close()

	repo, err := keeper.NewSQLite(db)
	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}

	defer repo.Close()

	if err := database.Seed(ctx, db, repo, reg); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	proxyService := proxy.New(repo)

	cli.New(repo, proxyService).Run()
}
