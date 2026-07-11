package auth

import (
	"context"
	"sync"
)

// MemoryUserStore implements UserStore in memory for testing.
type MemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users: make(map[string]*User),
	}
}

func (s *MemoryUserStore) CreateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.Email]; exists {
		return ErrUserExists
	}
	s.users[user.Email] = user
	return nil
}

func (s *MemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

var (
	ErrUserExists   = UserError("user already exists")
	ErrUserNotFound = UserError("user not found")
)

type UserError string

func (e UserError) Error() string { return string(e) }
