package markdown

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

var (
	headerRegex   = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	backtickRegex = regexp.MustCompile("`([^`]+)`")

	// Modality Matchers (ordered by specificity — must-not checked before must)
	mustNotRegex   = regexp.MustCompile(`(?i)\b(must not|shall not|cannot|never|may not)\b`)
	mustRegex      = regexp.MustCompile(`(?i)\b(must|requires?|shall|enforces?|has to|have to)\b`)
	shouldRegex    = regexp.MustCompile(`(?i)\b(should|recommended|ought to)\b`)
	idempotentWord = regexp.MustCompile(`(?i)\b(idempotent|idempotency)\b`)

	// Normative section names — claims are only extracted from these
	normativeSectionKeywords = []string{
		"decision", "consequence", "context", "specification",
		"invariant", "requirement", "rule", "policy", "constraint",
		"approach", "design", "contract", "behavior", "behaviour",
	}

	// Bullet list prefixes to strip before parsing
	bulletPrefixes = []string{"- ", "* ", "+ ", "• "}
)

// ParseADR parses an Architecture Decision Record into structured claims.
func ParseADR(filePath string, tenantID uuid.UUID, workspace string) ([]knowledge.ClaimIR, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var claims []knowledge.ClaimIR
	scanner := bufio.NewScanner(file)
	// Increase buffer for long lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	docTitle := filePath
	docTitleSet := false
	currentSection := "Preamble"
	lineNumber := 0
	inCodeBlock := false

	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Toggle code block state
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Track Markdown sections
		if matches := headerRegex.FindStringSubmatch(line); len(matches) == 3 {
			level := len(matches[1])
			text := strings.TrimSpace(matches[2])
			if level == 1 && !docTitleSet {
				docTitle = text
				docTitleSet = true
			}
			currentSection = text
			continue
		}

		// Only parse sections likely to assert normative invariants
		if !isNormativeSection(currentSection) {
			continue
		}

		// Strip bullet prefix if present
		content := stripBulletPrefix(line)
		if content == "" {
			continue
		}

		// Skip blockquotes and meta-lines
		if strings.HasPrefix(content, ">") || strings.HasPrefix(content, "|") {
			continue
		}

		// Detect modality
		modality := detectModality(content)
		if modality == "" {
			continue
		}

		// Extract backticked symbols
		symbols := backtickRegex.FindAllStringSubmatch(content, -1)
		hasIdempotency := idempotentWord.MatchString(content)

		// -------------------------------------------------------------
		// Scenario A: Two or more backticked symbols
		// Example: "`RefundHandler` must call `PaymentGateway`"
		// -------------------------------------------------------------
		if len(symbols) >= 2 {
			subj := symbols[0][1]
			obj := symbols[1][1]
			pred := knowledge.PredicateCalls
			if modality == knowledge.ModalityMustNot {
				pred = knowledge.PredicateCalls
			}
			claims = append(claims, newClaim(
				tenantID, workspace, filePath, docTitle, currentSection,
				lineNumber, line, subj, modality, pred, obj,
				knowledge.ProvenanceExtracted, 0.95,
			))
			continue
		}

		// -------------------------------------------------------------
		// Scenario B: Single backticked symbol + idempotency keyword
		// Example: "`RefundHandler` must be idempotent"
		// -------------------------------------------------------------
		if len(symbols) == 1 && hasIdempotency {
			claims = append(claims, newClaim(
				tenantID, workspace, filePath, docTitle, currentSection,
				lineNumber, line, symbols[0][1], modality,
				knowledge.PredicateIdempotent, "IdempotencyKey",
				knowledge.ProvenanceExtracted, 0.90,
			))
			continue
		}

		// -------------------------------------------------------------
		// Scenario C: Natural language normative statement
		// Example: "The HTTP server MUST use the chi router."
		// (This was the missing case — the function existed but was never called)
		// -------------------------------------------------------------
		if len(symbols) == 0 {
			subj, pred, obj := extractNaturalLanguageClaim(content, modality)
			if subj != "" && obj != "" {
				claims = append(claims, newClaim(
					tenantID, workspace, filePath, docTitle, currentSection,
					lineNumber, line, subj, modality, pred, obj,
					knowledge.ProvenanceInferred, 0.75,
				))
			}
		}
	}

	return claims, scanner.Err()
}

// newClaim builds a ClaimIR with consistent defaults.
func newClaim(
	tenantID uuid.UUID,
	workspace, filePath, docTitle, section string,
	lineNumber int,
	rawLine, subject string,
	modality knowledge.Modality,
	predicate knowledge.Predicate,
	object string,
	provenance knowledge.ProvenanceClass,
	confidence float64,
) knowledge.ClaimIR {
	return knowledge.ClaimIR{
		ID:            uuid.New(),
		Workspace:     workspace,
		TenantID:      tenantID,
		DocumentPath:  filePath,
		DocumentTitle: docTitle,
		DocumentType:  "ADR",
		SectionTitle:  section,
		LineStart:     lineNumber,
		LineEnd:       lineNumber,
		RawStatement:  rawLine,
		Subject:       subject,
		Modality:      modality,
		Predicate:     predicate,
		Object:        object,
		Provenance:    provenance,
		Confidence:    confidence,
		Status:        knowledge.StatusUnverified,
	}
}

// isNormativeSection returns true if the section title suggests normative content.
func isNormativeSection(section string) bool {
	lower := strings.ToLower(section)
	for _, kw := range normativeSectionKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// stripBulletPrefix removes leading bullet markers from a line.
func stripBulletPrefix(line string) string {
	for _, prefix := range bulletPrefixes {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return line
}

// detectModality returns the strongest modality present in the line.
func detectModality(line string) knowledge.Modality {
	switch {
	case mustNotRegex.MatchString(line):
		return knowledge.ModalityMustNot
	case mustRegex.MatchString(line):
		return knowledge.ModalityMust
	case shouldRegex.MatchString(line):
		return knowledge.ModalityShould
	default:
		return ""
	}
}

// extractNaturalLanguageClaim parses a plain-English normative statement
// into (subject, predicate, object) using modal verb boundaries.
//
// Example: "The HTTP server MUST use the chi router."
//
//	→ subject = "HTTP server", predicate = "use", object = "chi router"
//
// Example: "The router MUST NOT use the default mux."
//
//	→ subject = "router", predicate = "use", object = "default mux"
func extractNaturalLanguageClaim(line string, modality knowledge.Modality) (string, knowledge.Predicate, string) {
	lowerLine := strings.ToLower(line)

	// Modal boundaries in priority order (longest first to avoid partial matches)
	modals := []string{
		"must not", "shall not", "cannot", "may not", "never",
		"must", "shall", "requires", "require", "enforces", "enforce",
		"has to", "have to",
		"should", "recommended", "ought to",
	}

	// Find earliest modal
	modalIdx := -1
	matchedModal := ""
	for _, m := range modals {
		if idx := strings.Index(lowerLine, m); idx >= 0 {
			if modalIdx == -1 || idx < modalIdx {
				modalIdx = idx
				matchedModal = m
			}
		}
	}

	if modalIdx <= 0 || matchedModal == "" {
		return "", "", ""
	}

	// Subject = everything before the modal
	subject := strings.TrimSpace(line[:modalIdx])
	// Strip common leading articles
	for _, prefix := range []string{"The ", "the ", "A ", "a ", "An ", "an ", "This ", "this ", "That ", "that "} {
		subject = strings.TrimPrefix(subject, prefix)
	}
	// Strip trailing colon (section-like prefixes)
	subject = strings.TrimSuffix(subject, ":")
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "", "", ""
	}

	// Rest = everything after the modal
	rest := strings.TrimSpace(line[modalIdx+len(matchedModal):])
	// Remove common leading verbs of obligation
	for _, prefix := range []string{"be ", "have ", "not ", "always ", "never "} {
		if strings.HasPrefix(strings.ToLower(rest), prefix) {
			// Don't strip "not" if it's the only content
			if prefix == "not " && len(rest) <= 4 {
				break
			}
			rest = strings.TrimSpace(rest[len(prefix):])
		}
	}
	parts := strings.Fields(rest)
	if len(parts) < 2 {
		return "", "", ""
	}

	// First word after modal = verb → predicate
	verb := strings.ToLower(parts[0])
	predicate := knowledge.Predicate(verb)

	// Rest = object
	object := strings.Join(parts[1:], " ")
	object = strings.TrimSuffix(object, ".")
	object = strings.TrimSuffix(object, ",")
	object = strings.TrimSuffix(object, ";")
	object = strings.TrimSpace(object)

	if object == "" {
		return "", "", ""
	}

	return subject, predicate, object
}

// / ParseADRPublic wraps ParseADR for use with normalized content.
// It reads from contentPath but stores sourcePath as the claim's origin.
func ParseADRPublic(contentPath, sourcePath string, tenantID uuid.UUID, workspace string) ([]knowledge.ClaimIR, error) {
	claims, err := ParseADR(contentPath, tenantID, workspace)
	if err != nil {
		return nil, err
	}
	// Rewrite DocumentPath to reflect the original source file
	for i := range claims {
		claims[i].DocumentPath = sourcePath
	}
	return claims, nil
}
