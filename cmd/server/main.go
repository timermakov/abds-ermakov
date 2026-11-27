package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/timermakov/abds-ermakov/internal/config"
	"github.com/timermakov/abds-ermakov/internal/server"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	db, err := connect(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
	}()
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: server.New(cfg, db),
	}
	if err := httpServer.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func connect(cfg config.Config) (*sql.DB, error) {
	options := &clickhouse.Options{
		Addr: []string{cfg.ClickHouseHost + ":" + cfg.ClickHousePort},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDatabase,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePassword,
		},
	}
	var lastErr error
	for attempt := 1; attempt <= cfg.ConnectRetries; attempt++ {
		db := clickhouse.OpenDB(options)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectPingTimeout)
		err := db.PingContext(ctx)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = err
		_ = db.Close()
		if attempt < cfg.ConnectRetries {
			time.Sleep(cfg.ConnectRetryDelay)
		}
	}
	return nil, fmt.Errorf("connect to clickhouse after %d attempts: %w", cfg.ConnectRetries, lastErr)
}
