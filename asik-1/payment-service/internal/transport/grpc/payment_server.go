package grpc

import (
	"context"
	"errors"

	paymentv1 "github.com/usenbai-nur/ADP2_asik2_generated/payment/v1"

	"github.com/ap2/payment-service/internal/domain"
	"github.com/ap2/payment-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	paymentUseCase *usecase.PaymentUseCase
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{paymentUseCase: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	output, err := s.paymentUseCase.Authorize(ctx, usecase.AuthorizeInput{
		OrderID:       req.GetOrderId(),
		Amount:        req.GetAmount(),
		CustomerEmail: "user@example.com",
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrderID), errors.Is(err, domain.ErrInvalidAmount):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to process payment")
		}
	}

	return toPaymentResponse(output.Payment), nil
}

func (s *PaymentServer) GetPaymentByOrderID(ctx context.Context, req *paymentv1.GetPaymentByOrderIDRequest) (*paymentv1.PaymentResponse, error) {
	payment, err := s.paymentUseCase.GetByOrderID(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return nil, status.Error(codes.NotFound, "payment not found")
		}
		return nil, status.Error(codes.Internal, "failed to fetch payment")
	}

	return toPaymentResponse(payment), nil
}

func toPaymentResponse(p *domain.Payment) *paymentv1.PaymentResponse {
	return &paymentv1.PaymentResponse{
		Id:            p.ID,
		OrderId:       p.OrderID,
		TransactionId: p.TransactionID,
		Amount:        p.Amount,
		Status:        toProtoPaymentStatus(p.Status),
		CreatedAt:     timestamppb.New(p.CreatedAt),
	}
}

func toProtoPaymentStatus(statusValue string) paymentv1.PaymentStatus {
	switch statusValue {
	case domain.StatusAuthorized:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_AUTHORIZED
	case domain.StatusDeclined:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_DECLINED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}