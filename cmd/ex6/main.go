package main

import (
	"context"
	"fmt"
	"log/slog"
	"postgres-workshop/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderResult struct {
	ID           int64   `db:"id"`
	CustomerName string  `db:"customer_name"`
	Status       string  `db:"status"`
	Total        float64 `db:"total"`
}

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		slog.Error("не удалось создать пул соединений", slog.Any("error", err))
		return
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		slog.Error("не удалось подключиться к базе данных", slog.Any("error", err))
		return
	}
	slog.Info("Подключение установлено")

	demoUnionInjection(ctx, pool)
}

func demoDeleteBlocked(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("=== Демо 1: DELETE-инъекция через Query (extended protocol) ===")

	countOrders(ctx, pool, "ДО атаки")

	slog.Info("--- Нормальный ввод: Alice ---")
	searchOrdersViaQuery(ctx, pool, "Alice")

	slog.Info("--- АТАКА: '; DELETE FROM orders; --' ---")
	slog.Info("pgx использует extended protocol → multi-statement ЗАБЛОКИРОВАН")
	searchOrdersViaQuery(ctx, pool, "'; DELETE FROM orders; --")

	countOrders(ctx, pool, "ПОСЛЕ атаки (данные на месте)")
}

func demoDeleteViaExec(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("=== Демо 2: DELETE-инъекция через Exec (simple protocol) ===")

	countOrders(ctx, pool, "ДО атаки")
	
	slog.Info("--- Нормальный ввод: Alice ---")
	updateStatusUnsafe(ctx, pool, "Alice")

	countOrders(ctx, pool, "ПОСЛЕ нормального запроса (данные на месте)")

	slog.Info("--- АТАКА: '; DELETE FROM orders; --' ---")
	updateStatusUnsafe(ctx, pool, "'; DELETE FROM orders; --")

	countOrders(ctx, pool, "ПОСЛЕ атаки")

	slog.Info("")
	slog.Info("Для восстановления: cd cmd/example07_sql_injection && task migrate:down && task migrate:up")
}


func updateStatusUnsafe(ctx context.Context, pool *pgxpool.Pool, customerName string) {
	query := "UPDATE orders SET status = 'viewed' WHERE customer_name = '" + customerName + "'"
	slog.Info("Выполняем запрос", slog.String("sql", query))

	_, err := pool.Exec(ctx, query)
	if err != nil {
		slog.Error("Ошибка Exec", slog.Any("error", err))
	}
}

func demoUnionInjection(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("=== Демо 3: UNION-инъекция (кража адресов доставки) ===")
	slog.Info("Extended protocol НЕ защищает — это один SQL-запрос, не два")

	slog.Info("--- Нормальный ввод: Alice ---")
	searchOrdersViaQuery(ctx, pool, "Alice")

	slog.Info("--- АТАКА: UNION SELECT с адресами и телефонами ---")
	payload := "' UNION SELECT id, delivery_address, phone, total FROM orders --"
	searchOrdersViaQuery(ctx, pool, payload)
}

func searchOrdersViaQuery(ctx context.Context, pool *pgxpool.Pool, customerName string) {
	query := "SELECT id, customer_name, status, total FROM orders WHERE customer_name = '" + customerName + "'"
	slog.Info("Выполняем запрос", slog.String("sql", query))

	rows, err := pool.Query(ctx, query)
	if err != nil {
		slog.Error("Ошибка запроса", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[OrderResult])
	if err != nil {
		slog.Error("Ошибка CollectRows", slog.Any("error", err))
		return
	}

	for _, o := range orders {
		slog.Info("Результат",
			slog.Int64("id", o.ID),
			slog.String("customer_name", o.CustomerName),
			slog.String("status", o.Status),
			slog.Float64("total", o.Total),
		)
	}

	slog.Info(fmt.Sprintf("Найдено строк: %d", len(orders)))
}

func countOrders(ctx context.Context, pool *pgxpool.Pool, label string) {
	var count int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM orders").Scan(&count)
	if err != nil {
		slog.Error("Ошибка подсчёта", slog.Any("error", err))
		return
	}
	slog.Info(fmt.Sprintf("Заказов в таблице (%s): %d", label, count))
}
