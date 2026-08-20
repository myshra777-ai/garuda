// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package profile

type AccountProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (p *AccountProfile) DisplayName() string {
	return "User: " + p.Email
}

func FormatEmail(email string) string {
	return email
}
