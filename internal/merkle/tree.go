package merkle

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

// Domain separation prefixes (RFC 6962)
const (
	leafPrefix     byte = 0x00
	internalPrefix byte = 0x01
)

// LeafCommitment computes the Merkle leaf commitment for raw leaf data.
func LeafCommitment(data []byte) []byte {
	h := sha256.New()
	h.Write([]byte{leafPrefix})
	h.Write(data)
	return h.Sum(nil)
}

// InternalCommitment computes a Merkle internal node commitment.
func InternalCommitment(left, right []byte) []byte {
	h := sha256.New()
	h.Write([]byte{internalPrefix})
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}

// ProofNode represents one step in an inclusion proof.
type ProofNode struct {
	Position byte   // 'L' or 'R' — where the sibling sits relative to current hash
	Hash     []byte // 32 bytes
}

const (
	SiblingLeft  byte = 'L'
	SiblingRight byte = 'R'
)

var (
	ErrEmptyTree       = errors.New("merkle: cannot build proof for empty tree")
	ErrIndexOutOfRange = errors.New("merkle: leaf index out of range")
)

// BuildRoot computes the Merkle root for the given leaves.
// Leaves are hashed with the leaf prefix; internal nodes with internal prefix.
// Empty leaf set returns SHA256(0x01 || "GARUDA_EMPTY_TREE_V1") as a defined
// constant — this is a valid but empty-tree root, not a random hash.
func BuildRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		h := sha256.New()
		h.Write([]byte{internalPrefix})
		h.Write([]byte("GARUDA_EMPTY_TREE_V1"))
		return h.Sum(nil)
	}

	level := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		level[i] = LeafCommitment(leaf)
	}

	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				next = append(next, InternalCommitment(level[i], level[i+1]))
			} else {
				// Odd leaf promoted unchanged
				next = append(next, level[i])
			}
		}
		level = next
	}

	return level[0]
}

// BuildProof constructs an inclusion proof for leaf at index.
func BuildProof(leaves [][]byte, index int) ([]ProofNode, error) {
	if len(leaves) == 0 {
		return nil, ErrEmptyTree
	}
	if index < 0 || index >= len(leaves) {
		return nil, fmt.Errorf("%w: index=%d, count=%d", ErrIndexOutOfRange, index, len(leaves))
	}

	level := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		level[i] = LeafCommitment(leaf)
	}

	var proof []ProofNode
	idx := index

	for len(level) > 1 {
		var sibling []byte
		var position byte

		if idx%2 == 0 {
			// Current node is on the left; sibling (if any) is on the right
			if idx+1 < len(level) {
				sibling = level[idx+1]
				position = SiblingRight
			} else {
				// Odd node — no sibling, promote unchanged
				// Add a proof node indicating "no sibling"
				// Actually: for verification simplicity, we don't add anything
				// if there's no sibling, because the promotion is deterministic
			}
		} else {
			// Current node is on the right; sibling is on the left
			sibling = level[idx-1]
			position = SiblingLeft
		}

		if sibling != nil {
			proof = append(proof, ProofNode{Position: position, Hash: sibling})
		}

		// Build next level
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				next = append(next, InternalCommitment(level[i], level[i+1]))
			} else {
				next = append(next, level[i])
			}
		}
		level = next
		idx = idx / 2
	}

	return proof, nil
}

// VerifyInclusion verifies that leaf data is included in a tree with the given root.
// Pure function: takes only the leaf data, proof, and root. No DB, no chain replay.
func VerifyInclusion(leafData []byte, proof []ProofNode, root []byte) bool {
	current := LeafCommitment(leafData)

	for _, node := range proof {
		var combined []byte
		switch node.Position {
		case SiblingLeft:
			combined = InternalCommitment(node.Hash, current)
		case SiblingRight:
			combined = InternalCommitment(current, node.Hash)
		default:
			return false
		}
		current = combined
	}

	return bytesEqual(current, root)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	// Constant-time compare
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
