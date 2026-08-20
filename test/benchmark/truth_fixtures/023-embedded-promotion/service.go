// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package service

type BaseStore struct {
	connectionString string
}

func (b *BaseStore) Query(q string) []string {
	return []string{"result-1", "result-2"}
}

type OrderService struct {
	BaseStore
	TenantID string
}

func (s *OrderService) FetchOrders() []string {
	return s.Query("SELECT * FROM orders")
}
