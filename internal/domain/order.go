package domain

import "time"

type Order struct {
	ID           int64     `db:"id"`
	CustomerName string    `db:"customer_name"`
	Status       string    `db:"status"`
	Total        float64   `db:"total"`
	CreatedAt    time.Time `db:"created_at"`
}
