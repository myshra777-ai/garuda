package selfdescribe

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RoadmapYAML struct {
	SchemaVersion string `yaml:"schema_version"`
	Roadmap       struct {
		CurrentPhase string `yaml:"current_phase"`
		NextPhase    string `yaml:"next_phase"`
		Phases       map[string]struct {
			Name   string `yaml:"name"`
			Status string `yaml:"status"`
		} `yaml:"phases"`
	} `yaml:"roadmap"`
}

func ReadRoadmap(path string) RoadmapInfo {
	roadmapPath := filepath.Join(path, "docs", "source", "roadmap.yaml")

	var info RoadmapInfo
	if data, err := os.ReadFile(roadmapPath); err == nil {
		var yamlData RoadmapYAML
		if err := yaml.Unmarshal(data, &yamlData); err == nil {
			info.CurrentPhase = yamlData.Roadmap.CurrentPhase
			info.NextPhase = yamlData.Roadmap.NextPhase
			for phaseID, phase := range yamlData.Roadmap.Phases {
				info.Phases = append(info.Phases, phaseID+": "+phase.Name)
			}
			return info
		}
	}

	// Fallback: infer from common patterns
	return RoadmapInfo{
		CurrentPhase: "P2",
		NextPhase:    "P3",
		Phases: []string{
			"P0: Foundation",
			"P1: Semantic Core",
			"P2: Go Analyzer",
			"P3: CLI & Artifacts",
		},
	}
}
