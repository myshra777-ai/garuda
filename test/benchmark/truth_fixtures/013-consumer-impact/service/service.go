package service

import (
	"context"
	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/013-consumer-impact/store"
)

type UserService struct {
	store store.UserStore
}

func NewUserService(s store.UserStore) *UserService {
	return &UserService{store: s}
}

func (s *UserService) RegisterUser(ctx context.Context, id, email string) error {
	return s.store.SaveUser(ctx, store.UserRecord{ID: id, Email: email})
}
