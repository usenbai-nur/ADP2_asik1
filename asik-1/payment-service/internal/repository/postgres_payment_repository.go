package repository

import (
	"context"
	"database/sql"

	"github.com/ap2/payment-service/internal/domain"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Save inserts a new payment record.
func (r *PostgresPaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
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
	query := `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments WHERE order_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
