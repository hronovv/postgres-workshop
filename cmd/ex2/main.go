package main

import (
	"context"
	"fmt"
	"log/slog"
	"postgres-workshop/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
)

type Order struct {
	ID           int64     `db:"id"`
	CustomerName string    `db:"customer_name"`
	Status       string    `db:"status"`
	Total        float64   `db:"total"`
	CreatedAt    time.Time `db:"created_at"`
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.DSN) // создается лишь 1 соединение(TCP)
	/*
	 Дорого на каждый запрос создавать соединение, а шарить соединение между горутинами = DATA RACE
	*/
	if err != nil {
		slog.Error("pgx.Connect", slog.Any("error", err))
	}

	defer conn.Close(ctx)

	demoCollectRows(ctx, conn)
}

func demoCollectRows(ctx context.Context, conn *pgx.Conn) {
	rows, err := conn.Query(ctx, "SELECT id, customer_name, status, total, created_at FROM orders;")
	if err != nil {
		slog.Error("Query", slog.Any("error", err))
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		slog.Error("CollectRows", slog.Any("error", err))
	}

	for _, order := range orders {
		fmt.Println(order)
	}
}
