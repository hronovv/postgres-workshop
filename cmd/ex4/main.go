package main

import (
	"context"
	"fmt"
	"log/slog"
	"postgres-workshop/internal/config"
	"postgres-workshop/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		slog.Error("pgxpool.New", slog.Any("error", err))
		return
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("pool.Ping", slog.Any("error", err))
		return
	}
	slog.Info("Подключение к базе данных произошло успешно")

	newID := demoInsert(ctx, pool)
	if newID == 0 {
		return
	}

	demoSelectOne(ctx, pool, newID)
	demoSelectList(ctx, pool)
	demoUpdate(ctx, pool, newID)
	demoDelete(ctx, pool, newID)
}

func demoInsert(ctx context.Context, pool *pgxpool.Pool) int64 {
	var newID int64
	err := pool.QueryRow(ctx,
		`
		INSERT INTO ORDERS (customer_name, status, total)
		VALUES
			($1,$2,$3)
			RETURNING id;
		`, "Ваня Ванечкин", "new", 4545.23,
	).Scan(&newID)
	if err != nil {
		slog.Error("INSERT ERROR", slog.Any("error", err))
		return 0
	}
	slog.Info("Создан заказ", slog.Int64("id", newID))
	return newID
}

func demoSelectOne(ctx context.Context, pool *pgxpool.Pool, id int64) {
	rows, err := pool.Query(ctx,
		`
		SELECT
			id,
			customer_name,
			status, total,
			created_at
		FROM orders
		WHERE
			id = $1;
		`, id,
	)
	if err != nil {
		slog.Error("SELECT ERROR", slog.Any("error", err))
		return
	}
	order, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Order])
	if err != nil {
		slog.Error("error CollectOneRow", slog.Any("error", err))
		return
	}

	slog.Info("Получен заказ",
		slog.Int64("id", order.ID),
		slog.String("покупатель", order.CustomerName),
		slog.String("статус", order.Status),
		slog.Float64("сумма", order.Total),
	)
}

func demoSelectList(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx,
		` SELECT id, customer_name, status, total, created_at
		FROM orders
		ORDER BY created_at DESC;
		`,
	)
	if err != nil {
		slog.Error("SELECT ERROR", slog.Any("error", err))
		return
	}
	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Order])
	if err != nil {
		slog.Error("error CollectRows", slog.Any("error", err))
		return
	}

	slog.Info("got list of orders", slog.Int("всего", len(orders)))
	for _, order := range orders {
		fmt.Println(order)
	}
}

func demoUpdate(ctx context.Context, pool *pgxpool.Pool, id int64) {
	tag, err := pool.Exec(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2;`,
		"paid", id,
	)
	if err != nil {
		slog.Error("error UPDATE", slog.Any("error", err))
		return
	}

	slog.Info("UPDATED successfully", slog.Int64("affected rows", tag.RowsAffected()))
}

func demoDelete(ctx context.Context, pool *pgxpool.Pool, id int64) {
	tag, err := pool.Exec(ctx,
		`DELETE FROM orders WHERE id=$1;`,
		id,
	)

	if err != nil {
		slog.Error("error DELETE", slog.Any("error", err))
		return
	}

	slog.Info("DELETED successfully", slog.Int64("affected rows", tag.RowsAffected()))
}
