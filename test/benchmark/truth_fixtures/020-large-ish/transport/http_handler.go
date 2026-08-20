package transport

import (
	"context"
	"errors"

	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/020-large-ish/domain"
	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/020-large-ish/service"
)

type HTTPHandler struct {
	orderSvc *service.OrderService
}

func NewHTTPHandler(svc *service.OrderService) *HTTPHandler {
	return &HTTPHandler{orderSvc: svc}
}

func (h *HTTPHandler) HandleCreateOrder(ctx context.Context, userID string, amount int64) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("invalid order amount")
	}
	return h.orderSvc.PlaceOrder(userID, amount)
}
