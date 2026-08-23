// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

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
