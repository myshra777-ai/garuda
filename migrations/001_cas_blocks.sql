CREATE TABLE IF NOT EXISTS cas_blocks (
    block_hash BYTEA PRIMARY KEY,
    content TEXT NOT NULL,
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cas_created ON cas_blocks(created_at);