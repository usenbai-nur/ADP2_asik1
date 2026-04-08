package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ap2/order-service/internal/domain"
)

type HTTPPaymentClient struct {
	baseURL    string
	httpClient *http.Client // shared, pre-configured with a 2-second timeout
}

func NewHTTPPaymentClient(baseURL string, httpClient *http.Client) *HTTPPaymentClient {
	return &HTTPPaymentClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

type authorizeRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type authorizeResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"` // "Authorized" | "Declined"
	Message       string `json:"message,omitempty"`
}

// Authorize sends POST /payments to the Payment Service and returns the outcome.
func (c *HTTPPaymentClient) Authorize(ctx context.Context, orderID string, amount int64) (*domain.PaymentResult, error) {
	payload, err := json.Marshal(authorizeRequest{OrderID: orderID, Amount: amount})
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Covers: connection refused, timeout, DNS failure.
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}

	var out authorizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode payment response: %w", err)
	}

	return &domain.PaymentResult{
		TransactionID: out.TransactionID,
		Status:        out.Status,
	}, nil
}
