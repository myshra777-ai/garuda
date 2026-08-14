package analyzer

// Import only what is needed for this file.
// If you don't use github.com/google/uuid in this file, remove it.
// The types are defined in types.go.

// Analyze runs the full analysis pipeline on a given root path
func Analyze(root string) (*Result, error) {
	result, err := Extract(root)
	if err != nil {
		return nil, err
	}
	// Additional post-processing could go here
	return result, nil
}

// (No Provenance or RevisionSummary definitions here – they are in types.go)
// Result is an alias for Result to maintain compatibility with the store.
