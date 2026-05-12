package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/ap2/order-service/internal/domain"
	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
	publisher     domain.OrderStatusPublisher
	cache         domain.OrderCache
}

func NewOrderUseCase(
	repo domain.OrderRepository,
	paymentClient domain.PaymentClient,
	publisher domain.OrderStatusPublisher,
	cache domain.OrderCache,
) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
		publisher:     publisher,
		cache:         cache,
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
		if err != nil && !errors.Is(err, domain.ErrOrderNotFound) {
			return nil, err
		}
	}

	now := time.Now().UTC()

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     input.CustomerID,
		ItemName:       input.ItemName,
		Amount:         input.Amount,
		Status:         domain.StatusPending,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		if errors.Is(err, domain.ErrDuplicateRequest) && input.IdempotencyKey != "" {
			existing, lookupErr := uc.repo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
			if lookupErr == nil && existing != nil {
				return &CreateOrderOutput{Order: existing}, nil
			}
		}
		return nil, err
	}

	result, err := uc.paymentClient.Authorize(ctx, order.ID, order.Amount)
	if err != nil {
		updatedAt := time.Now().UTC()
		if updateErr := uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed, updatedAt); updateErr != nil {
			return nil, updateErr
		}

		uc.invalidateCache(ctx, order.ID)

		order.Status = domain.StatusFailed
		order.UpdatedAt = updatedAt

		uc.publisher.Publish(domain.OrderStatusEvent{
			OrderID:   order.ID,
			Status:    order.Status,
			UpdatedAt: order.UpdatedAt,
		})

		return nil, domain.ErrPaymentUnavailable
	}

	newStatus := domain.StatusFailed
	if result.Status == "Authorized" {
		newStatus = domain.StatusPaid
	}

	updatedAt := time.Now().UTC()
	if err := uc.repo.UpdateStatus(ctx, order.ID, newStatus, updatedAt); err != nil {
		return nil, err
	}

	uc.invalidateCache(ctx, order.ID)

	order.Status = newStatus
	order.UpdatedAt = updatedAt

	uc.publisher.Publish(domain.OrderStatusEvent{
		OrderID:   order.ID,
		Status:    order.Status,
		UpdatedAt: order.UpdatedAt,
	})

	return &CreateOrderOutput{Order: order}, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if uc.cache != nil {
		cachedOrder, err := uc.cache.Get(ctx, id)
		if err == nil {
			return cachedOrder, nil
		}
	}

	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}

	if uc.cache != nil {
		_ = uc.cache.Set(ctx, order)
	}

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}

	if !order.CanBeCancelled() {
		return nil, domain.ErrCannotCancel
	}

	updatedAt := time.Now().UTC()
	if err := uc.repo.UpdateStatus(ctx, id, domain.StatusCancelled, updatedAt); err != nil {
		return nil, err
	}

	uc.invalidateCache(ctx, id)

	order.Status = domain.StatusCancelled
	order.UpdatedAt = updatedAt

	uc.publisher.Publish(domain.OrderStatusEvent{
		OrderID:   order.ID,
		Status:    order.Status,
		UpdatedAt: order.UpdatedAt,
	})

	return order, nil
}

func (uc *OrderUseCase) GetOrderStats(ctx context.Context) (map[string]int, error) {
	return uc.repo.CountByStatus(ctx)
}

func (uc *OrderUseCase) invalidateCache(ctx context.Context, orderID string) {
	if uc.cache != nil {
		_ = uc.cache.Delete(ctx, orderID)
	}
}