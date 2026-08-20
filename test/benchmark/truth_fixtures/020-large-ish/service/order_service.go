// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"time"

	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/020-large-ish/domain"
)

type OrderService struct {
	userRepo  domain.UserRepository
	orderRepo domain.OrderRepository
}

func NewOrderService(u domain.UserRepository, o domain.OrderRepository) *OrderService {
	return &OrderService{
		userRepo:  u,
		orderRepo: o,
	}
}

func (s *OrderService) PlaceOrder(userID string, total int64) (*domain.Order, error) {
	user, err := s.userRepo.GetUser(userID)
	if err != nil || user == nil {
		return nil, errors.New("cannot place order for non-existent user")
	}

	order := &domain.Order{
		ID:        "ord_12345",
		UserID:    user.ID,
		Total:     total,
		Status:    "PENDING",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.orderRepo.SaveOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}
