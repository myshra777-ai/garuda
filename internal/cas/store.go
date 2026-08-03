package cas

import (
	"context"

	"github.com/myshra777-ai/garuda/internal/types"
)

// BlockStore defines the interface for content‑addressable block storage.
type BlockStore interface {
	// IngestBlocks performs a synchronous, idempotent upsert of blocks.
	// If a block already exists, its ref_count is NOT changed (only content is preserved).
	IngestBlocks(ctx context.Context, blocks []types.Block) error
}
