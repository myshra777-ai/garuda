package api

import (
	"net/http"
	"strings"
)

// getModelInfo extracts model name from headers or request context and determines the provider.
func getModelInfo(r *http.Request) (modelName, modelProvider string) {
	modelName = r.Header.Get("X-Model")
	if modelName == "" {
		if val, ok := r.Context().Value("model_name").(string); ok && val != "" {
			modelName = val
		} else {
			modelName = "unknown"
		}
	}

	modelProvider = r.Header.Get("X-Model-Provider")
	if modelProvider == "" {
		if val, ok := r.Context().Value("model_provider").(string); ok && val != "" {
			modelProvider = val
		}
	}

	if modelProvider == "" {
		lowerModel := strings.ToLower(modelName)
		switch {
		case strings.HasPrefix(lowerModel, "gpt"):
			modelProvider = "openai"
		case strings.HasPrefix(lowerModel, "claude"):
			modelProvider = "anthropic"
		case strings.HasPrefix(lowerModel, "gemini"):
			modelProvider = "google"
		case strings.HasPrefix(lowerModel, "deepseek"):
			modelProvider = "deepseek"
		default:
			modelProvider = "local"
		}
	}

	return modelName, modelProvider
}
