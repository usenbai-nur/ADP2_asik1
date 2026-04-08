package domain

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
	ErrInvalidCustomerID  = errors.New("customer_id is required")
	ErrInvalidItemName    = errors.New("item_name is required")
	ErrCannotCancel       = errors.New("only pending orders can be cancelled")
	ErrDuplicateRequest   = errors.New("duplicate request: order already exists for this idempotency key")
	ErrPaymentUnavailable = errors.New("payment service unavailable")
)
