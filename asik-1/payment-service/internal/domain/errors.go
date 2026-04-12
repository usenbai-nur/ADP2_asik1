package domain

import "errors"

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidOrderID  = errors.New("order_id is required")
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
)