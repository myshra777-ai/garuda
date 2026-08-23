package payments

type UserPaymentRequest struct {
	UserID    string
	Amount    float64
	Currency  string
	CardToken string
}

type PaymentProcessor interface {
	ProcessPayment(req UserPaymentRequest) (string, error)
	RefundPayment(txID string) error
}

func CalculateFee(amount float64) float64 {
	return amount * 0.029
}
