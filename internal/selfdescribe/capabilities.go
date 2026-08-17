package selfdescribe

import (
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// ExtractCapabilities extracts capabilities from the analyzer result.
// Every entity (struct, interface, function) is considered a capability.
// Functions are mapped to commands if they match CLI patterns.
func ExtractCapabilities(result *analyzer.Result) []Capability {
	var caps []Capability

	// Track seen names to avoid duplicates
	seen := make(map[string]bool)

	// Extract from entities (structs, interfaces, functions)
	for _, entity := range result.Entities {
		// Skip common internal types that aren't user-facing
		if isInternal(entity.Name) {
			continue
		}

		// Map entity kind to a capability
		cap := Capability{
			Name:   entity.Name,
			Status: "stable",
			Source: entity.File,
		}

		// If it's a function that looks like a command, link it
		if strings.HasPrefix(entity.Name, "Handle") || strings.HasPrefix(entity.Name, "Cmd") {
			cap.Command = strings.ToLower(strings.TrimPrefix(entity.Name, "Handle"))
			cap.Command = strings.TrimPrefix(cap.Command, "Cmd")
			cap.Command = strings.ToLower(cap.Command)
		}

		// If it's a struct that looks like a command group
		if strings.HasSuffix(entity.Name, "Cmd") || strings.HasSuffix(entity.Name, "Command") {
			cap.Command = strings.ToLower(strings.TrimSuffix(entity.Name, "Cmd"))
			cap.Command = strings.TrimSuffix(cap.Command, "Command")
		}

		// Add description if available
		if entity.Comments != "" {
			cap.Description = entity.Comments
		} else {
			cap.Description = "Provides " + entity.Name + " functionality"
		}

		key := cap.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		caps = append(caps, cap)
	}

	// Add planned capabilities from roadmap patterns (if found in codebase)
	// We'll infer from comments like "// TODO: implement X" or "// Planned: X"
	// For now, this is a placeholder; actual roadmap is read from YAML.

	return caps
}

// isInternal skips entities that are clearly internal/private
func isInternal(name string) bool {
	internalPrefixes := []string{
		"internal", "private", "_", "test", "mock", "fake", "stub", "gen", "generated",
	}
	for _, p := range internalPrefixes {
		if strings.HasPrefix(strings.ToLower(name), p) {
			return true
		}
	}
	return false
}
