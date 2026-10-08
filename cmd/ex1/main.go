package main

import (
	"database/sql"
	"log/slog"
	"postgres-workshop/internal/config"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // просто регистрируем драйвер, у него там init()
)

func main() {
	db, err := sql.Open("pgx", config.DSN)
	if err != nil {
		slog.Error("sql.Open", slog.Any("error", err))
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		slog.Error("db.Ping", slog.Any("error", err))
		return
	}
	slog.Info("Соединение установлено")

	row := db.QueryRow("SELECT id, customer_name, status, total, created_at FROM orders WHERE id=2;")
	var (
		id           int64
		customerName string
		status       string
		total        float64
		createdAt    time.Time
	)
	if err := row.Scan(&id, &customerName, &status, &total, &createdAt); err != nil {
		slog.Error("row.Scan", slog.Any("error", err))
		return
	}
	slog.Info("Получен заказ",
		slog.Int64("id", id),
		slog.String("покупатель", customerName),
		slog.String("статус", status),
		slog.Float64("сумма", total),
		slog.String("создан", createdAt.Format("2006-01-02 15:04:03")),
	)

	stats := db.Stats()
	slog.Info("Статистика пула (database/sql)",
		slog.Int("открытых_соединений:", stats.OpenConnections),
		slog.Int("используется_сейчас:", stats.InUse),
		slog.Int("простаивает", stats.Idle),
		slog.Int64("ожидали_соединения", stats.WaitCount))
}
