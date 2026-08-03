package harvester

import (
	"regexp"
	"strings"
)

// Extractor extracts decisions from text using heuristics + optional LLM.
type Extractor struct {
	keywords      []string
	decisionRegex *regexp.Regexp
}

// NewExtractor creates a new Extractor.
func NewExtractor(keywords []string) *Extractor {
	// Build regex: (decided|we should|let's go with) ...
	pattern := `(?i)(` + strings.Join(keywords, "|") + `)\s*([^.!?]+[.!?])`
	re := regexp.MustCompile(pattern)

	return &Extractor{
		keywords:      keywords,
		decisionRegex: re,
	}
}

// Extract finds decision statements in text.
func (e *Extractor) Extract(text string) (string, float64) {
	matches := e.decisionRegex.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return "", 0.0
	}

	// Take the first match
	decision := strings.TrimSpace(matches[0][2])
	confidence := 0.7 + float64(len(matches))*0.05
	if confidence > 0.95 {
		confidence = 0.95
	}

	return decision, confidence
}

// ExtractWithContext combines previous extraction with thread context.
func (e *Extractor) ExtractWithContext(combinedText, previousDecision string) (string, float64) {
	// If we already have a decision, use it and boost confidence
	if previousDecision != "" {
		return previousDecision, 0.85
	}
	return e.Extract(combinedText)
}
