package api

import "github.com/myshra777-ai/garuda/internal/store"

func FakePaymentHandler() {
	// This is the violation: API layer directly calling the store layer
	_, _ = store.NewPostgresStore("test")
}
