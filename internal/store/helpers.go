package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectModulePath reads go.mod and extracts the module path
func DetectModulePath(repoPath string) (string, error) {
	modPath := filepath.Join(repoPath, "go.mod")
	f, err := os.Open(modPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}
