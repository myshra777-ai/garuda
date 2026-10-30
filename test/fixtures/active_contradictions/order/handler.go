// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"fmt"
	"net/http"
)

type DirectDB struct{}

func (d *DirectDB) RawQuery(query string) error {
	return nil
}

type OrderHandler struct {
	db *DirectDB
}

func (h *OrderHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Processing checkout request...")

	// ❌ ACTIVE ARCHITECTURAL CONTRADICTION:
	// Violates ADR-0004 §2: "OrderHandler MUST NOT call DirectDB.RawQuery"
	// While unit tests and syntax linters pass, Garuda AST code graph detects
	// this forbidden call edge and halts the deployment.
	_ = h.db.RawQuery("SELECT * FROM orders WHERE status = 'pending'")
}

