package app

import (
	"sync"

	"github.com/ap2/order-service/internal/domain"
)

type OrderStatusHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan domain.OrderStatusEvent]struct{}
}

func NewOrderStatusHub() *OrderStatusHub {
	return &OrderStatusHub{
		subscribers: make(map[string]map[chan domain.OrderStatusEvent]struct{}),
	}
}

func (h *OrderStatusHub) Subscribe(orderID string) (<-chan domain.OrderStatusEvent, func()) {
	ch := make(chan domain.OrderStatusEvent, 8)

	h.mu.Lock()
	if _, ok := h.subscribers[orderID]; !ok {
		h.subscribers[orderID] = make(map[chan domain.OrderStatusEvent]struct{})
	}
	h.subscribers[orderID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		subs, ok := h.subscribers[orderID]
		if !ok {
			return
		}

		if _, exists := subs[ch]; exists {
			delete(subs, ch)
			close(ch)
		}

		if len(subs) == 0 {
			delete(h.subscribers, orderID)
		}
	}

	return ch, unsubscribe
}

func (h *OrderStatusHub) Publish(event domain.OrderStatusEvent) {
	h.mu.RLock()
	subs := h.subscribers[event.OrderID]
	channels := make([]chan domain.OrderStatusEvent, 0, len(subs))
	for ch := range subs {
		channels = append(channels, ch)
	}
	h.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
		}
	}
}