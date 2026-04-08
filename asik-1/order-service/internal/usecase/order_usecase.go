package usecase

import (
	"context"
	"time"

	"github.com/ap2/order-service/internal/domain"
	"github.com/google/uuid"
)

// OrderUseCase contains all business rules for the order bounded context.
type OrderUseCase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
}

func NewOrderUseCase(repo domain.OrderRepository, paymentClient domain.PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

type CreateOrderOutput struct {
	Order *domain.Order
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error) {
	if input.IdempotencyKey != "" {
		existing, err := uc.repo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil && existing != nil {
			return &CreateOrderOutput{Order: existing}, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     input.CustomerID,
		ItemName:       input.ItemName,
		Amount:         input.Amount,
		Status:         domain.StatusPending,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := order.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	result, err := uc.paymentClient.Authorize(ctx, order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed)
		order.Status = domain.StatusFailed
		return nil, domain.ErrPaymentUnavailable
	}

	newStatus := domain.StatusFailed
	if result.Status == "Authorized" {
		newStatus = domain.StatusPaid
	}
	if err := uc.repo.UpdateStatus(ctx, order.ID, newStatus); err != nil {
		return nil, err
	}
	order.Status = newStatus

	return &CreateOrderOutput{Order: order}, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

// CancelOrder enforces the invariant: only Pending orders can be cancelled.
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	if !order.CanBeCancelled() {
		return nil, domain.ErrCannotCancel
	}
	if err := uc.repo.UpdateStatus(ctx, id, domain.StatusCancelled); err != nil {
		return nil, err
	}
	order.Status = domain.StatusCancelled
	return order, nil
}



// defence task
// GetOrderStats
func (uc *OrderUseCase) GetOrderStats(ctx context.Context) (map[string]int, error) {
	return uc.repo.CountByStatus(ctx)
}