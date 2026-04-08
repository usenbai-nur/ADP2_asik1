package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ap2/order-service/internal/domain"
)

// PostgresOrderRepository is the concrete persistence adapter for orders.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new repository wired to the given *sql.DB.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Save inserts a new order row.
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		nullableString(order.IdempotencyKey),
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key,''), created_at, updated_at
		FROM orders WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	if key == "" {
		return nil, domain.ErrOrderNotFound
	}
	query := `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key,''), created_at, updated_at
		FROM orders WHERE idempotency_key = $1
	`
	row := r.db.QueryRowContext(ctx, query, key)
	return scanOrder(row)
}

// UpdateStatus changes the status of an existing order.
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), id)
	return err
}

// CountByStatus UUUUUUUUU
func (r *PostgresOrderRepository) CountByStatus(ctx context.Context) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM orders GROUP BY status`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]int{
		"pending":   0,
		"paid":      0,
		"failed":    0,
		"cancelled": 0,
		"total":     0,
	}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		switch status {
		case "Pending":
			result["pending"] = count
		case "Paid":
			result["paid"] = count
		case "Failed":
			result["failed"] = count
		case "Cancelled":
			result["cancelled"] = count
		}
		result["total"] += count
	}
	return result, nil
}

func scanOrder(row *sql.Row) (*domain.Order, error) {
	var o domain.Order
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.ItemName,
		&o.Amount,
		&o.Status,
		&o.IdempotencyKey,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}