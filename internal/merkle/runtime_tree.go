package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type RuntimeLeaf struct {
	SourceID      string
	TargetID      string
	Status        string
	ObservedCount int64
	Reason        string
	RawTarget     string
	LastTraceID   string
}

// ComputeLeafHash generates a deterministic SHA-256 digest for a single claim verification leaf.
func (l *RuntimeLeaf) ComputeLeafHash() string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s|%s|%s|%d|%s|%s|%s",
		l.SourceID,
		l.TargetID,
		l.Status,
		l.ObservedCount,
		l.Reason,
		l.RawTarget,
		l.LastTraceID,
	)))
	return hex.EncodeToString(hasher.Sum(nil))
}

// BuildRuntimeMerkleRoot builds a deterministic Merkle root from runtime verification leaves.
func BuildRuntimeMerkleRoot(leaves []RuntimeLeaf) (string, int) {
	if len(leaves) == 0 {
		emptyHash := sha256.Sum256([]byte("GARUDA_EMPTY_RUNTIME_TREE"))
		return hex.EncodeToString(emptyHash[:]), 0
	}

	hashes := make([]string, len(leaves))
	for i, leaf := range leaves {
		hashes[i] = leaf.ComputeLeafHash()
	}

	// Lexicographical sort guarantees canonical root regardless of SQL query ordering
	sort.Strings(hashes)

	for len(hashes) > 1 {
		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := sha256.Sum256([]byte(hashes[i] + hashes[i+1]))
				nextLevel = append(nextLevel, hex.EncodeToString(combined[:]))
			} else {
				// Promote odd leaf
				combined := sha256.Sum256([]byte(hashes[i] + hashes[i]))
				nextLevel = append(nextLevel, hex.EncodeToString(combined[:]))
			}
		}
		hashes = nextLevel
	}

	return hashes[0], len(leaves)
}

// ComputeUnifiedEpochRoot combines static AST root and runtime verification root into an epoch hash.
func ComputeUnifiedEpochRoot(staticRoot, runtimeRoot string, parentHash string, height int64) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%s:%d", staticRoot, runtimeRoot, parentHash, height)))
	return hex.EncodeToString(hasher.Sum(nil))
}
