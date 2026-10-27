// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultDBURL is the local development Postgres connection string.
const DefaultDBURL = "postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable"

// FindDotEnvPath returns the path to the nearest .env file, or ".env" if none exists yet.
func FindDotEnvPath() string {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, ".env"),
			filepath.Join(filepath.Dir(wd), ".env"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
		)
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ".env"
}

// LoadDotEnv reads key=value pairs from .env and sets any environment variables
// that are not already set in the current process environment.
func LoadDotEnv() {
	path := FindDotEnvPath()
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// UpsertDotEnv updates or appends keys in the .env file so configuration
// persists across terminal sessions and restarts automatically.
func UpsertDotEnv(pairs map[string]string) error {
	path := FindDotEnvPath()
	existingLines := []string{}
	seenKeys := map[string]bool{}

	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				if idx := strings.IndexByte(trimmed, '='); idx > 0 {
					k := strings.TrimSpace(trimmed[:idx])
					if newVal, ok := pairs[k]; ok {
						existingLines = append(existingLines, fmt.Sprintf("%s=%s", k, newVal))
						seenKeys[k] = true
						continue
					}
				}
			}
			existingLines = append(existingLines, line)
		}
	}

	for k, v := range pairs {
		if !seenKeys[k] {
			existingLines = append(existingLines, fmt.Sprintf("%s=%s", k, v))
		}
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}

	out := strings.TrimRight(strings.Join(existingLines, "\n"), "\n") + "\n"
	return os.WriteFile(path, []byte(out), 0600)
}

// ResolveDBURL returns DATABASE_URL from env/.env or DefaultDBURL.
func ResolveDBURL() string {
	LoadDotEnv()
	if u := strings.TrimSpace(os.Getenv("DATABASE_URL")); u != "" {
		return strings.ReplaceAll(u, "@localhost:", "@127.0.0.1:")
	}
	return DefaultDBURL
}

// ResolveAPIAddr returns the base HTTP URL for the Garuda API gateway,
// respecting GARUDA_API_URL, GARUDA_API_BASE, and PORT.
func ResolveAPIAddr() string {
	LoadDotEnv()
	if u := strings.TrimSpace(os.Getenv("GARUDA_API_URL")); u != "" {
		return strings.TrimRight(u, "/")
	}
	if u := strings.TrimSpace(os.Getenv("GARUDA_API_BASE")); u != "" {
		return strings.TrimRight(u, "/")
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	return "http://localhost:" + port
}
