package domain

import "time"

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64 // cents
	Status        string
	CreatedAt     time.Time
}

// Payment limit: orders above 100 000 cents are declined.
const MaxAuthorizedAmount int64 = 100_000

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"
)
