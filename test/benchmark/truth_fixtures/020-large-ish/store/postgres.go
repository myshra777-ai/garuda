// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"database/sql"
	"errors"

	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/020-large-ish/domain"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) GetUser(id string) (*domain.User, error) {
	if id == "" {
		return nil, errors.New("user not found")
	}
	return &domain.User{ID: id, Email: "user@example.com"}, nil
}

func (s *PostgresStore) CreateUser(u *domain.User) error {
	return nil
}

func (s *PostgresStore) SaveOrder(o *domain.Order) error {
	return nil
}

func (s *PostgresStore) GetOrderByID(id string) (*domain.Order, error) {
	return &domain.Order{ID: id, Total: 100, Status: "CONFIRMED"}, nil
}
