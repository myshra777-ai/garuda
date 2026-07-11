CREATE TABLE cas_blocks (
    block_hash BYTEA PRIMARY KEY,
    content TEXT NOT NULL,
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cas_created ON cas_blocks(created_at);