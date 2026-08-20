// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package payments

import "context"

type UserPaymentRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
}

type PaymentProcessor interface {
	Process(ctx context.Context, accountID string, amount int64) (bool, error)
}

func CalculateFee(amount int64, tier string) int64 {
	if tier == "vip" {
		return 0
	}
	return amount / 100
}

func ChargeCustomer(ctx context.Context, p PaymentProcessor, req UserPaymentRequest) error {
	_, err := p.Process(ctx, req.AccountID, req.Amount)
	return err
}
