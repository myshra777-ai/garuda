package payments

type UserPaymentRequest struct {
	UserID   string
	Amount   float64
	Currency string
}

type PaymentProcessor interface {
	ProcessPayment(req UserPaymentRequest) (string, error)
}

func CalculateFee(amount float64, tier string) float64 {
	return amount * 0.025
}
