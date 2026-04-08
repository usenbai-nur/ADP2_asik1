package domain

import "time"

// Order represents an order in the online store.
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

// Allowed order statuses.
const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

// Validate enforces order invariants.
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

// CanBeCancelled returns true only when the order is still Pending.
func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusPending
}
