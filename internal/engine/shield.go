package engine

import (
	"regexp"
)

type ShieldEngine struct {
	patterns []*regexp.Regexp
}

func NewShieldEngine() *ShieldEngine {
	return &ShieldEngine{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(api_key|apikey|secret|password|bearer|auth_token)\s*=\s*['"][a-zA-Z0-9_\-]{16,}['"]`),
			regexp.MustCompile(`(sk-[a-zA-Z0-9]{32,})`),
			regexp.MustCompile(`(ghp_[a-zA-Z0-9]{36})`),
			regexp.MustCompile(`(xox[baprs]-[a-zA-Z0-9]{10,})`),
			regexp.MustCompile(`(?i)\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), // Email/PII
		},
	}
}

// Redact sanitizes sensitive data inline prior to provider dispatch
func (s *ShieldEngine) Redact(input string) (string, int) {
	cleaned := input
	count := 0

	for _, pattern := range s.patterns {
		matches := pattern.FindAllString(cleaned, -1)
		count += len(matches)
		cleaned = pattern.ReplaceAllString(cleaned, "[REDACTED_SECRET]")
	}

	return cleaned, count
}
