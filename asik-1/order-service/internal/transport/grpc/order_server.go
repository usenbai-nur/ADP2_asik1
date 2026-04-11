package grpc

import (
	"context"
	"errors"
	"io"
	"time"

	orderv1 "github.com/usenbai-nur/ADP2_asik2_generated/order/v1"

	"github.com/ap2/order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type orderRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Order, error)
}

type orderStatusSubscriber interface {
	Subscribe(orderID string) (<-chan domain.OrderStatusEvent, func())
}

type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	repo orderRepository
	hub  orderStatusSubscriber
}

func NewOrderServer(repo orderRepository, hub orderStatusSubscriber) *OrderServer {
	return &OrderServer{
		repo: repo,
		hub:  hub,
	}
}

func (s *OrderServer) SubscribeToOrderUpdates(
	req *orderv1.OrderRequest,
	stream orderv1.OrderService_SubscribeToOrderUpdatesServer,
) error {
	orderID := req.GetOrderId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.repo.FindByID(stream.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return status.Error(codes.NotFound, "order not found")
		}
		return status.Error(codes.Internal, "failed to load order")
	}

	if err := stream.Send(toProtoUpdate(order.ID, order.Status, order.UpdatedAt)); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return status.Error(codes.Internal, "failed to send initial order status")
	}

	events, unsubscribe := s.hub.Subscribe(orderID)
	defer unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return nil
			}

			if err := stream.Send(toProtoUpdate(event.OrderID, event.Status, event.UpdatedAt)); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return status.Error(codes.Internal, "failed to send order update")
			}
		}
	}
}

func toProtoUpdate(orderID, statusValue string, updatedAt time.Time) *orderv1.OrderStatusUpdate {
	return &orderv1.OrderStatusUpdate{
		OrderId:   orderID,
		Status:    toProtoOrderStatus(statusValue),
		UpdatedAt: timestamppb.New(updatedAt),
	}
}

func toProtoOrderStatus(statusValue string) orderv1.OrderStatus {
	switch statusValue {
	case domain.StatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case domain.StatusPaid:
		return orderv1.OrderStatus_ORDER_STATUS_PAID
	case domain.StatusFailed:
		return orderv1.OrderStatus_ORDER_STATUS_FAILED
	case domain.StatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}