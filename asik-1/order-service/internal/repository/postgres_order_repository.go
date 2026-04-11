package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ap2/order-service/internal/domain"
	"github.com/lib/pq"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	const query = `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		nullableString(order.IdempotencyKey),
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "orders_idempotency_key_key" {
			return domain.ErrDuplicateRequest
		}
		return err
	}
	return nil
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	const query = `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	if key == "" {
		return nil, domain.ErrOrderNotFound
	}

	const query = `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at, updated_at
		FROM orders
		WHERE idempotency_key = $1
	`

	row := r.db.QueryRowContext(ctx, query, key)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error {
	const query = `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, updatedAt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *PostgresOrderRepository) CountByStatus(ctx context.Context) (map[string]int, error) {
	const query = `SELECT status, COUNT(*) FROM orders GROUP BY status`

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
		case domain.StatusPending:
			result["pending"] = count
		case domain.StatusPaid:
			result["paid"] = count
		case domain.StatusFailed:
			result["failed"] = count
		case domain.StatusCancelled:
			result["cancelled"] = count
		}
		result["total"] += count
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}