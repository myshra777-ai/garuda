package selfdescribe

import "time"

// SelfDescription is the complete output of garuda self describe
type SelfDescription struct {
	SchemaVersion string        `json:"schema_version"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Source        SourceInfo    `json:"source"`
	Product       ProductInfo   `json:"product"`
	Capabilities  []Capability  `json:"capabilities"`
	CLI           CLIInfo       `json:"cli"`
	Semantic      SemanticInfo  `json:"semantic_model"`
	Benchmarks    BenchmarkInfo `json:"benchmarks"`
	Trust         TrustInfo     `json:"trust"`
	Roadmap       RoadmapInfo   `json:"roadmap"`
}

type SourceInfo struct {
	Repository    string   `json:"repository"`
	Commit        string   `json:"commit"`
	Branch        string   `json:"branch"`
	Language      string   `json:"language"`
	LanguageScope []string `json:"language_scope"`
}

type ProductInfo struct {
	Name        string   `json:"name"`
	Tagline     string   `json:"tagline"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Thesis      string   `json:"thesis"`
	Audiences   []string `json:"audiences"`
}

type Capability struct {
	Name        string `json:"name"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description"`
	Status      string `json:"status"` // stable, beta, planned
	Source      string `json:"source"` // file path or evidence
	Phase       string `json:"phase,omitempty"`
}

type CLIInfo struct {
	Commands []CLICommand `json:"commands"`
}

type CLICommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Flags       []string `json:"flags,omitempty"`
}

type SemanticInfo struct {
	Entities      int  `json:"entities"`
	Relationships int  `json:"relationships"`
	Evidence      int  `json:"evidence"`
	Lineage       bool `json:"lineage"`
}

type BenchmarkInfo struct {
	Available bool                   `json:"available"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

type TrustInfo struct {
	ImmutableLedger bool   `json:"immutable_ledger"`
	MerkleRoot      string `json:"merkle_root"`
	RevisionCount   int    `json:"revision_count"`
	AuditTrail      bool   `json:"audit_trail"`
}

type RoadmapInfo struct {
	CurrentPhase string   `json:"current_phase"`
	NextPhase    string   `json:"next_phase"`
	Phases       []string `json:"phases"`
}
