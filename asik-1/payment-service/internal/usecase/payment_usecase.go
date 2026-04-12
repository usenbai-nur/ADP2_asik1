package usecase

import (
	"context"
	"time"

	"github.com/ap2/payment-service/internal/domain"
	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

type AuthorizeInput struct {
	OrderID string
	Amount  int64
}

type AuthorizeOutput struct {
	Payment *domain.Payment
}

func (uc *PaymentUseCase) Authorize(ctx context.Context, input AuthorizeInput) (*AuthorizeOutput, error) {
	if input.OrderID == "" {
		return nil, domain.ErrInvalidOrderID
	}
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	status := domain.StatusAuthorized
	if input.Amount > domain.MaxAuthorizedAmount {
		status = domain.StatusDeclined
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       input.OrderID,
		TransactionID: uuid.New().String(),
		Amount:        input.Amount,
		Status:        status,
		CreatedAt:     time.Now().UTC(),
	}

	if err := uc.repo.Save(ctx, payment); err != nil {
		return nil, err
	}

	return &AuthorizeOutput{Payment: payment}, nil
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, domain.ErrPaymentNotFound
	}
	return payment, nil
}