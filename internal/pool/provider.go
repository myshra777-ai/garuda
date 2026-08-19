// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProviderClient defines the interface for each LLM provider.
type ProviderClient interface {
	// HealthCheck verifies the API key is valid and returns latency.
	HealthCheck(ctx context.Context, apiKey string) (*HealthResult, error)
	// Name returns the provider name.
	Name() string
	// BaseURL returns the API base URL.
	BaseURL() string
}

// HealthResult contains the result of a health check.
type HealthResult struct {
	Valid     bool       `json:"valid"`
	LatencyMs int64      `json:"latency_ms"`
	Error     string     `json:"error,omitempty"`
	Quota     *QuotaInfo `json:"quota,omitempty"`
}

// QuotaInfo contains rate limit information.
type QuotaInfo struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`
}

// GeminiClient implements ProviderClient for Google Gemini.
type GeminiClient struct{}

func (g *GeminiClient) Name() string    { return "gemini" }
func (g *GeminiClient) BaseURL() string { return "https://generativelanguage.googleapis.com" }

func (g *GeminiClient) HealthCheck(ctx context.Context, apiKey string) (*HealthResult, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/v1beta/models?key=%s", g.BaseURL(), apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     "invalid API key",
		}, nil
	}
	if resp.StatusCode == 429 {
		return &HealthResult{
			Valid:     true,
			LatencyMs: latency,
			Error:     "rate limited",
			Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
		}, nil
	}
	if resp.StatusCode != 200 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("HTTP %d", resp.StatusCode),
		}, nil
	}

	var body struct {
		Models []interface{} `json:"models"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &body)

	return &HealthResult{
		Valid:     true,
		LatencyMs: latency,
		Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
	}, nil
}

// DeepSeekClient implements ProviderClient for DeepSeek.
type DeepSeekClient struct{}

func (d *DeepSeekClient) Name() string    { return "deepseek" }
func (d *DeepSeekClient) BaseURL() string { return "https://api.deepseek.com" }

func (d *DeepSeekClient) HealthCheck(ctx context.Context, apiKey string) (*HealthResult, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/v1/models", d.BaseURL())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     "invalid API key",
		}, nil
	}
	if resp.StatusCode == 429 {
		return &HealthResult{
			Valid:     true,
			LatencyMs: latency,
			Error:     "rate limited",
			Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
		}, nil
	}
	if resp.StatusCode != 200 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("HTTP %d", resp.StatusCode),
		}, nil
	}

	return &HealthResult{
		Valid:     true,
		LatencyMs: latency,
		Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
	}, nil
}

// OpenAIClient implements ProviderClient for OpenAI.
type OpenAIClient struct{}

func (o *OpenAIClient) Name() string    { return "openai" }
func (o *OpenAIClient) BaseURL() string { return "https://api.openai.com" }

func (o *OpenAIClient) HealthCheck(ctx context.Context, apiKey string) (*HealthResult, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/v1/models", o.BaseURL())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     "invalid API key",
		}, nil
	}
	if resp.StatusCode == 429 {
		return &HealthResult{
			Valid:     true,
			LatencyMs: latency,
			Error:     "rate limited",
			Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
		}, nil
	}
	if resp.StatusCode != 200 {
		return &HealthResult{
			Valid:     false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("HTTP %d", resp.StatusCode),
		}, nil
	}

	return &HealthResult{
		Valid:     true,
		LatencyMs: latency,
		Quota:     &QuotaInfo{RequestsPerMinute: 60, TokensPerMinute: 1000000},
	}, nil
}

// NewProviderClient returns the appropriate client for a provider.
func NewProviderClient(provider string) (ProviderClient, error) {
	switch provider {
	case "gemini":
		return &GeminiClient{}, nil
	case "deepseek":
		return &DeepSeekClient{}, nil
	case "openai":
		return &OpenAIClient{}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
