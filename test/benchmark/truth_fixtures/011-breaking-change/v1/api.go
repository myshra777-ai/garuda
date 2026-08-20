// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package payments

import "context"

type UserPaymentRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type PaymentProcessor interface {
	Process(accountID string, amount int64) bool
	Refund(transactionID string) error
}

func CalculateFee(amount int64) int64 {
	return amount / 100
}

func ChargeCustomer(ctx context.Context, p PaymentProcessor, req UserPaymentRequest) error {
	_ = p.Process(req.AccountID, req.Amount)
	return nil
}
