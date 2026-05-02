package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ap2/payment-service/internal/domain"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	const query = `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		p.ID,
		p.OrderID,
		p.TransactionID,
		p.Amount,
		p.Status,
		p.CreatedAt,
	)
	return err
}

func (r *PostgresPaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, orderID)

	var p domain.Payment
	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.TransactionID,
		&p.Amount,
		&p.Status,
		&p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PostgresPaymentRepository) GetStats(ctx context.Context) (*domain.PaymentStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = 'Authorized') AS authorized_count,
			COUNT(*) FILTER (WHERE status = 'Declined') AS declined_count,
			COALESCE(SUM(amount), 0) AS total_amount
		FROM payments
	`

	var stats domain.PaymentStats

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalCount,
		&stats.AuthorizedCount,
		&stats.DeclinedCount,
		&stats.TotalAmount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}