// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package selfdescribe

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProductYAML is the structure of docs/source/product.yaml
type ProductYAML struct {
	SchemaVersion string `yaml:"schema_version"`
	Product       struct {
		Name        string `yaml:"name"`
		DisplayName string `yaml:"display_name"`
		Tagline     string `yaml:"tagline"`
		Category    struct {
			Primary   string   `yaml:"primary"`
			Secondary []string `yaml:"secondary"`
		} `yaml:"category"`
		ShortDescription string `yaml:"short_description"`
		LongDescription  string `yaml:"long_description"`
		Thesis           string `yaml:"thesis"`
		ImmediateMission string `yaml:"immediate_mission"`
		Audiences        struct {
			Primary   []string `yaml:"primary"`
			Secondary []string `yaml:"secondary"`
		} `yaml:"audiences"`
		Positioning struct {
			Problem         string   `yaml:"problem"`
			Solution        string   `yaml:"solution"`
			Differentiation []string `yaml:"differentiation"`
		} `yaml:"positioning"`
	} `yaml:"product"`
}

// ReadProductMetadata reads product.yaml if it exists, otherwise infers from repo.
func ReadProductMetadata(path string) ProductInfo {
	productPath := filepath.Join(path, "docs", "source", "product.yaml")

	var yamlData ProductYAML
	if data, err := os.ReadFile(productPath); err == nil {
		if err := yaml.Unmarshal(data, &yamlData); err == nil {
			// Build audiences list
			audiences := yamlData.Product.Audiences.Primary
			audiences = append(audiences, yamlData.Product.Audiences.Secondary...)

			return ProductInfo{
				Name:        yamlData.Product.DisplayName,
				Tagline:     yamlData.Product.Tagline,
				Category:    yamlData.Product.Category.Primary,
				Description: yamlData.Product.ShortDescription,
				Thesis:      yamlData.Product.Thesis,
				Audiences:   audiences,
			}
		}
	}

	// Fallback: infer from repository name
	repoName := filepath.Base(path)
	if repoName == "." || repoName == "" {
		repoName = "garuda"
	}

	return ProductInfo{
		Name:        repoName,
		Tagline:     "A Go-based project",
		Category:    "Unknown",
		Description: "This project is automatically described by Garuda.",
		Thesis:      "Built with Go.",
		Audiences:   []string{"Developers"},
	}
}
