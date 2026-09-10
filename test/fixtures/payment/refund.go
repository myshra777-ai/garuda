package payment

type IdempotencyStore struct{}

func (s *IdempotencyStore) Record(key string) error {
	return nil
}

type DirectBankAPI struct{}

func (b *DirectBankAPI) Transfer() error {
	return nil
}

type RefundHandler struct {
	store *IdempotencyStore
	bank  *DirectBankAPI
}

func (h *RefundHandler) ProcessRefund() {
	// Satisfies claim 1: RefundHandler MUST CALL IdempotencyStore
	h.store.Record("test-key")

	// Violates claim 3: RefundHandler MUST_NOT CALL DirectBankAPI
	h.bank.Transfer()
}
