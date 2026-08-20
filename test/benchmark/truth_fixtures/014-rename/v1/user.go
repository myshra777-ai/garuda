// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package auth

type UserSession struct {
	Token     string
	ExpiresAt int64
}

func (s *UserSession) IsActive() bool {
	return s.ExpiresAt > 0
}
