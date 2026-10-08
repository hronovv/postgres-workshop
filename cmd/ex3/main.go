package main

import (
	"context"
	"fmt"
	"log/slog"
	"postgres-workshop/internal/config"
	"postgres-workshop/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		slog.Error("pgxpool.ParseConfig", slog.Any("error", err))
		return
	}

	config.MaxConns = 10
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.MinConns = 2
	//setting up config

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("pgxpool.NewWithConfig", slog.Any("error", err))
		return
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("pool.Ping", slog.Any("error", err))
		return
	}

	slog.Info("Пул соединений создан")

	rows, err := pool.Query(ctx, "SELECT id, customer_name, status, total, created_at FROM orders;")
	if err != nil {
		slog.Error("Query", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Order])
	if err != nil {
		slog.Error("CollectRows", slog.Any("error", err))
		return
	}

	for _, order := range orders {
		fmt.Println(order)
	}

	stat := pool.Stat()
	slog.Info("Статистика пула (pgxpool)",
		slog.Int("всего_соединений", int(stat.TotalConns())),
		slog.Int("простаивает", int(stat.IdleConns())),
		slog.Int("используется", int(stat.AcquiredConns())),
		slog.Int("максимум", int(stat.MaxConns())),
		slog.Int("конструируется_сейчас", int(stat.ConstructingConns())),
	)
}
