package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ap2/order-service/internal/domain"
	"github.com/ap2/order-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockOrderUseCase struct {
	CreateOrderFunc func(ctx interface{}, input usecase.CreateOrderInput) (*usecase.CreateOrderOutput, error)
}

func (m *mockOrderUseCase) CreateOrder(ctx interface{}, input usecase.CreateOrderInput) (*usecase.CreateOrderOutput, error) {
	return m.CreateOrderFunc(ctx, input)
}

func TestCreateOrder_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	mockUC := &mockOrderUseCase{
		CreateOrderFunc: func(ctx interface{}, input usecase.CreateOrderInput) (*usecase.CreateOrderOutput, error) {
			return nil, domain.ErrInvalidAmount
		},
	}
	h := NewOrderHandler(mockUC)
	h.RegisterRoutes(r)

	w := httptest.NewRecorder()
	body := `{"customer_id": "", "item_name": "", "amount": 0}`
	req, _ := http.NewRequest("POST", "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
