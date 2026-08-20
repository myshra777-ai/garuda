package auditing

import (
	"crypto/sha256"
	"encoding/hex"
)

// AuditLogger records security and compliance events with hashing.
type AuditLogger struct {
	Prefix string
}

// ComputeHash generates a SHA-256 digest over input event data.
func (l *AuditLogger) ComputeHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// LogSecurityEvent writes a formatted security audit record.
func (l *AuditLogger) LogSecurityEvent(actor, action string) (string, error) {
	record := l.Prefix + ":" + actor + ":" + action
	hash := l.ComputeHash([]byte(record))
	return hash, nil
}
