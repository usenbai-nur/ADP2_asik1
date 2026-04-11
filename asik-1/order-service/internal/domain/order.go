package domain

import "time"

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if o.ItemName == "" {
		return ErrInvalidItemName
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusPending
}

type OrderStatusEvent struct {
	OrderID   string
	Status    string
	UpdatedAt time.Time
}