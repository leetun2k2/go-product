package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-bedrock/postgresx"
	"github.com/leetun2k2/go-bedrock/redisx"
	"github.com/leetun2k2/go-bedrock/restapix"
	"github.com/leetun2k2/go-bedrock/serverx"
	"github.com/leetun2k2/go-product/config"
	"github.com/leetun2k2/go-product/internal/transport"
)

func run(_ []string) error {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Initialize logger
	logger, err := logx.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	// 3. Create root context
	ctx := context.Background()
	logger.Info(ctx, "logger initialized", "config", cfg.Logger)

	// 4. Initialize PostgreSQL (required runtime dependency)
	if cfg.Postgres == nil || cfg.Postgres.URL == "" {
		return errors.New("initialize postgres: configuration URL is required")
	}

	db, err := postgresx.New(ctx, *cfg.Postgres)
	if err != nil {
		return fmt.Errorf("initialize postgres: %w", err)
	}
	defer db.Close()

	// 5. Initialize Redis only when enabled
	var cache *redisx.Client
	if cfg.Redis != nil && cfg.Redis.Enabled {
		c, err := redisx.New(ctx, cfg.Redis.Config)
		if err != nil {
			return fmt.Errorf("initialize redis: %w", err)
		}
		cache = c
		defer func() {
			_ = cache.Close()
		}()
	}

	// 6. Initialize REST API and application
	api, err := restapix.New(*cfg.RestfulApi, logger)
	if err != nil {
		return fmt.Errorf("initialize restful api: %w", err)
	}

	app, err := transport.NewApplication(cfg, logger, db.Pool())
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	if err := app.RegisterRestfulApi(api.Engine()); err != nil {
		return fmt.Errorf("register restful api routes: %w", err)
	}

	// 7. Register components with serverx
	server := serverx.New(*cfg.Server, logger)
	if err := server.Register(db.Component()); err != nil {
		return fmt.Errorf("register postgres component: %w", err)
	}

	if cache != nil {
		if err := server.Register(cache.Component()); err != nil {
			return fmt.Errorf("register redis component: %w", err)
		}
	}

	if err := server.Register(api.Component()); err != nil {
		return fmt.Errorf("register restful api component: %w", err)
	}

	// 8. Run the server
	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
