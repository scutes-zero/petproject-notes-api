package main

import (
	"fmt"
	"log/slog"
	"notes-api/internal/app"
	"notes-api/internal/config"
	"notes-api/internal/storage"
	"notes-api/pkg/logger"
	"os"
)

// TODO: ADD MIGRATIONS?, ADD TESTS, CHECK HANDLERS
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.MustLoad()
	log := setupLogger()
	storage, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
		return err
	}

	app := app.NewApp(cfg, storage, log, []byte(cfg.JwtSecret))

	app.Start()
	return nil
}

func setupLogger() *slog.Logger {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))

	return logger
}
