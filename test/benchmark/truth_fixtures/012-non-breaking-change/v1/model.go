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
