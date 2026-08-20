package profile

import "strings"

type AccountProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Bio   string `json:"bio,omitempty"`
}

func (p *AccountProfile) DisplayName() string {
	if p.Bio != "" {
		return p.Email + " (" + p.Bio + ")"
	}
	return "User: " + p.Email
}

func (p *AccountProfile) IsVerified() bool {
	return p.Email != ""
}

func FormatEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
