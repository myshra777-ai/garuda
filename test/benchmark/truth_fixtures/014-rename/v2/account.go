package auth

// UserSession renamed to AccountSession (Preserves lineage).
type AccountSession struct {
	Token     string
	ExpiresAt int64
}

func (s *AccountSession) IsActive() bool {
	return s.ExpiresAt > 0
}
