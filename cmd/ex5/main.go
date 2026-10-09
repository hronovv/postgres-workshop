package main

import (
	"context"
	"log/slog"
	"postgres-workshop/internal/config"
	"postgres-workshop/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName       = "orders"
	colID           = "id"
	colCustomerName = "customer_name"
	colStatus       = "status"
	colTotal        = "total"
	colCreatedAt    = "created_at"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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
	slog.Info("Подключение к базе данных успешно установлено")

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
	insertSQL, insertArgs, err := psql.
		Insert(tableName).
		Columns(colCustomerName, colStatus, colTotal).
		Values("Пётр Петров", "new", 2500.00).
		Suffix("RETURNING " + colID).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения INSERT", slog.Any("error", err))
		return 0
	}
	slog.Info("Построен SQL INSERT", slog.String("sql", insertSQL), slog.Any("args", insertArgs))

	var newID int64
	err = pool.QueryRow(ctx, insertSQL, insertArgs...).Scan(&newID)
	if err != nil {
		slog.Error("ошибка выполнения INSERT", slog.Any("error", err))
		return 0
	}
	slog.Info("Создан заказ", slog.Int64("id", newID))

	return newID
}

func demoSelectOne(ctx context.Context, pool *pgxpool.Pool, id int64) {
	selectSQL, selectArgs, err := psql.
		Select(colID, colCustomerName, colStatus, colTotal, colCreatedAt).
		From(tableName).
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения SELECT", slog.Any("error", err))
		return
	}
	slog.Info("Построен SQL SELECT ONE", slog.String("sql", selectSQL), slog.Any("args", selectArgs))

	rows, err := pool.Query(ctx, selectSQL, selectArgs...)
	if err != nil {
		slog.Error("ошибка выполнения SELECT", slog.Any("error", err))
		return
	}

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Order])
	if err != nil {
		slog.Error("ошибка CollectOneRow", slog.Any("error", err))
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
	status := new("new")
	minTotal := new(100.0)

	qb := psql.
		Select(colID, colCustomerName, colStatus, colTotal, colCreatedAt).
		From(tableName).
		OrderBy(colCreatedAt + " DESC")

	if status != nil {
		qb = qb.Where(sq.Eq{colStatus: *status})
	}
	if minTotal != nil {
		qb = qb.Where(sq.GtOrEq{colTotal: *minTotal})
	}

	selectSQL, selectArgs, err := qb.ToSql()
	if err != nil {
		slog.Error("ошибка построения SELECT списка", slog.Any("error", err))
		return
	}
	slog.Info("Построен SQL SELECT LIST", slog.String("sql", selectSQL), slog.Any("args", selectArgs))

	rows, err := pool.Query(ctx, selectSQL, selectArgs...)
	if err != nil {
		slog.Error("ошибка выполнения SELECT списка", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Order])
	if err != nil {
		slog.Error("ошибка CollectRows", slog.Any("error", err))
		return
	}

	slog.Info("Найдено заказов", slog.Int("количество", len(orders)))
	for _, o := range orders {
		slog.Info("Заказ",
			slog.Int64("id", o.ID),
			slog.String("покупатель", o.CustomerName),
			slog.String("статус", o.Status),
			slog.Float64("сумма", o.Total),
		)
	}
}

func demoUpdate(ctx context.Context, pool *pgxpool.Pool, id int64) {
	updateSQL, updateArgs, err := psql.
		Update(tableName).
		Set(colStatus, "paid").
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения UPDATE", slog.Any("error", err))
		return
	}
	slog.Info("Построен SQL UPDATE", slog.String("sql", updateSQL), slog.Any("args", updateArgs))

	tag, err := pool.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		slog.Error("ошибка выполнения UPDATE", slog.Any("error", err))
		return
	}
	slog.Info("UPDATE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}

func demoDelete(ctx context.Context, pool *pgxpool.Pool, id int64) {
	deleteSQL, deleteArgs, err := psql.
		Delete(tableName).
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения DELETE", slog.Any("error", err))
		return
	}
	slog.Info("Построен SQL DELETE", slog.String("sql", deleteSQL), slog.Any("args", deleteArgs))

	tag, err := pool.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		slog.Error("ошибка выполнения DELETE", slog.Any("error", err))
		return
	}
	slog.Info("DELETE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}
