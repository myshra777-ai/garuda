package profile

// AccountProfile stores user profile details in v1.
type AccountProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// DisplayName returns a formatted display string in v1.
func (p *AccountProfile) DisplayName() string {
	return "User: " + p.Email
}

// FormatEmail cleans and lowercases email strings.
func FormatEmail(email string) string {
	return email
}
