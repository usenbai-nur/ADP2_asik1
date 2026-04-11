package repository

import (
	"context"
	"fmt"
	"time"

	paymentv1 "github.com/usenbai-nur/ADP2_asik2_generated/payment/v1"

	"github.com/ap2/order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCPaymentClient struct {
	client  paymentv1.PaymentServiceClient
	timeout time.Duration
}

func NewGRPCPaymentClient(client paymentv1.PaymentServiceClient, timeout time.Duration) *GRPCPaymentClient {
	return &GRPCPaymentClient{
		client:  client,
		timeout: timeout,
	}
}

func (c *GRPCPaymentClient) Authorize(ctx context.Context, orderID string, amount int64) (*domain.PaymentResult, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ProcessPayment(callCtx, &paymentv1.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.DeadlineExceeded, codes.Unavailable:
				return nil, domain.ErrPaymentUnavailable
			case codes.InvalidArgument:
				return nil, fmt.Errorf("payment validation failed: %w", err)
			default:
				return nil, fmt.Errorf("payment grpc call failed: %w", err)
			}
		}
		return nil, fmt.Errorf("payment grpc call failed: %w", err)
	}

	return &domain.PaymentResult{
		TransactionID: resp.GetTransactionId(),
		Status:        toDomainPaymentStatus(resp.GetStatus()),
	}, nil
}

func toDomainPaymentStatus(status paymentv1.PaymentStatus) string {
	switch status {
	case paymentv1.PaymentStatus_PAYMENT_STATUS_AUTHORIZED:
		return "Authorized"
	case paymentv1.PaymentStatus_PAYMENT_STATUS_DECLINED:
		return "Declined"
	default:
		return ""
	}
}