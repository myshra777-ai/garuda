package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkspaceConfig represents .garuda/workspace.yaml
type WorkspaceConfig struct {
	Workspace    string       `yaml:"workspace"`
	TenantID     string       `yaml:"tenant_id"`
	Repositories []RepoConfig `yaml:"repositories"`
	Documents    []DocConfig  `yaml:"documents"`
}

type RepoConfig struct {
	Path     string `yaml:"path"`
	Language string `yaml:"language,omitempty"`
	URL      string `yaml:"url,omitempty"`
}

type DocConfig struct {
	Path  string `yaml:"path"`
	Type  string `yaml:"type,omitempty"` // adr, openapi, policy, auto
	Watch bool   `yaml:"watch,omitempty"`
}

// LoadWorkspaceConfig reads and validates a config file.
func LoadWorkspaceConfig(path string) (*WorkspaceConfig, error) {
	path = ExpandHome(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Workspace == "" {
		return nil, fmt.Errorf("config missing required field: workspace")
	}
	if cfg.TenantID == "" {
		return nil, fmt.Errorf("config missing required field: tenant_id")
	}

	// Expand ~ in nested paths
	for i := range cfg.Repositories {
		cfg.Repositories[i].Path = ExpandHome(cfg.Repositories[i].Path)
	}
	for i := range cfg.Documents {
		cfg.Documents[i].Path = ExpandHome(cfg.Documents[i].Path)
	}

	return &cfg, nil
}

// FindWorkspaceConfig walks up from cwd looking for .garuda/workspace.yaml
func FindWorkspaceConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		candidate := filepath.Join(dir, ".garuda", "workspace.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .garuda/workspace.yaml found in %s or any parent directory", cwd)
}

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// ResolveDocPaths flattens configured doc sources into a list of concrete files.
// Directories are walked recursively; files are included directly.
func (cfg *WorkspaceConfig) ResolveDocPaths() ([]string, error) {
	var files []string

	for _, doc := range cfg.Documents {
		info, err := os.Stat(doc.Path)
		if err != nil {
			// Don't fail — just skip and warn
			fmt.Fprintf(os.Stderr, "⚠️  Skipping %s: %v\n", doc.Path, err)
			continue
		}

		if info.IsDir() {
			err := filepath.Walk(doc.Path, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return nil // skip unreadable entries
				}
				if f.IsDir() {
					base := filepath.Base(path)
					if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
						return filepath.SkipDir
					}
					return nil
				}
				ext := strings.ToLower(filepath.Ext(f.Name()))
				if isSupportedDocExt(ext) {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			files = append(files, doc.Path)
		}
	}

	return files, nil
}

func isSupportedDocExt(ext string) bool {
	switch ext {
	case ".md", ".markdown", ".pdf", ".docx", ".txt", ".rst":
		return true
	}
	return false
}
