package store

import "context"

type UserRecord struct {
	ID    string
	Email string
}

type UserStore interface {
	SaveUser(ctx context.Context, u UserRecord) error
}

type PostgresUserStore struct{}

func (s *PostgresUserStore) SaveUser(ctx context.Context, u UserRecord) error {
	return nil
}
