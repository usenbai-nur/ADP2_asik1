package domain

import "time"

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}

const MaxAuthorizedAmount int64 = 100_000

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"
)