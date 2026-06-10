package controller

import (
	"testing_go/auth"
	"testing_go/models"
)

// ─────────────────────────────────────────────────────────────────────────────
//  MODEL CAPABILITY REGISTRY
// ─────────────────────────────────────────────────────────────────────────────

// ModelCapability represents a supported model capability
type ModelCapability string

const (
	CapTextOnly   ModelCapability = "text"
	CapVision     ModelCapability = "vision"     // Accepts image input
	CapMultimodal ModelCapability = "multimodal"  // Accepts image + text + audio
	CapJSON       ModelCapability = "json"        // Supports response_format: json_object
	CapStreaming  ModelCapability = "streaming"   // Supports SSE streaming
)

// ModelInfo represents a discovered or known model
type ModelInfo struct {
	ID            string            `json:"id"`
	Provider      string            `json:"provider"`
	Capabilities  []ModelCapability `json:"capabilities"`
	ContextWindow int               `json:"context_window"`
	IsAvailable   bool              `json:"is_available"`
}

// knownModelCapabilities is a local registry of known models and their capabilities.
// Extend this map as new models become available.
var knownModelCapabilities = map[string]ModelInfo{
	// ── Nvidia NIM ──
	"meta/llama-3.1-70b-instruct":                {Provider: "nvidia", Capabilities: []ModelCapability{CapTextOnly, CapJSON}, ContextWindow: 128000},
	"meta/llama-3.1-8b-instruct":                 {Provider: "nvidia", Capabilities: []ModelCapability{CapTextOnly, CapJSON}, ContextWindow: 128000},
	"meta/llama-3.2-11b-vision-instruct":         {Provider: "nvidia", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 128000},
	"meta/llama-3.2-90b-vision-instruct":         {Provider: "nvidia", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 128000},
	"nvidia/llama-3.1-nemotron-70b-instruct":     {Provider: "nvidia", Capabilities: []ModelCapability{CapTextOnly, CapJSON}, ContextWindow: 128000},
	"google/gemma-2-27b-it":                      {Provider: "nvidia", Capabilities: []ModelCapability{CapTextOnly, CapJSON}, ContextWindow: 8192},
	"mistralai/mistral-large-2-instruct":         {Provider: "nvidia", Capabilities: []ModelCapability{CapTextOnly, CapJSON}, ContextWindow: 128000},
	"mistralai/mistral-small-3.1-24b-instruct":   {Provider: "nvidia", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 128000},

	// ── OpenAI ──
	"gpt-4o":      {Provider: "openai", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON, CapStreaming}, ContextWindow: 128000},
	"gpt-4o-mini": {Provider: "openai", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON, CapStreaming}, ContextWindow: 128000},
	"gpt-4-turbo": {Provider: "openai", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON, CapStreaming}, ContextWindow: 128000},
	"gpt-3.5-turbo": {Provider: "openai", Capabilities: []ModelCapability{CapTextOnly, CapJSON, CapStreaming}, ContextWindow: 16385},

	// ── Anthropic ──
	"claude-opus-4-20250514":   {Provider: "anthropic", Capabilities: []ModelCapability{CapVision, CapMultimodal}, ContextWindow: 200000},
	"claude-sonnet-4-20250514": {Provider: "anthropic", Capabilities: []ModelCapability{CapVision, CapMultimodal}, ContextWindow: 200000},
	"claude-3-5-sonnet-20240620": {Provider: "anthropic", Capabilities: []ModelCapability{CapVision, CapMultimodal}, ContextWindow: 200000},
	"claude-3-haiku-20240307":  {Provider: "anthropic", Capabilities: []ModelCapability{CapTextOnly}, ContextWindow: 200000},

	// ── Google Gemini ──
	"gemini-2.0-flash":      {Provider: "gemini", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 1048576},
	"gemini-2.0-flash-lite": {Provider: "gemini", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 1048576},
	"gemini-1.5-pro":        {Provider: "gemini", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 2097152},
	"gemini-1.5-flash":      {Provider: "gemini", Capabilities: []ModelCapability{CapVision, CapMultimodal, CapJSON}, ContextWindow: 1048576},
}

// FilterModelsByCapability returns models that have ALL of the requested capabilities
// and belong to providers for which the user has an active key.
func FilterModelsByCapability(user *models.User, requiredCaps []ModelCapability) []ModelInfo {
	availableProviders := userAvailableProviders(user)
	var result []ModelInfo
	for _, info := range knownModelCapabilities {
		if !containsProvider(availableProviders, info.Provider) {
			continue
		}
		if hasAllCapabilities(info.Capabilities, requiredCaps) {
			info.IsAvailable = true
			result = append(result, info)
		}
	}
	return result
}

// userAvailableProviders returns a list of providers for which the user has a decrypted API key.
func userAvailableProviders(user *models.User) []string {
	var providers []string
	if user.NvidiaKey != "" && auth.DecryptAPIKey(user.NvidiaKey) != "" {
		providers = append(providers, "nvidia")
	}
	if user.OpenAIKey != "" && auth.DecryptAPIKey(user.OpenAIKey) != "" {
		providers = append(providers, "openai")
	}
	if user.AnthropicKey != "" && auth.DecryptAPIKey(user.AnthropicKey) != "" {
		providers = append(providers, "anthropic")
	}
	if user.GeminiKey != "" && auth.DecryptAPIKey(user.GeminiKey) != "" {
		providers = append(providers, "gemini")
	}
	return providers
}

func containsProvider(providers []string, target string) bool {
	for _, p := range providers {
		if p == target {
			return true
		}
	}
	return false
}

func hasAllCapabilities(have, need []ModelCapability) bool {
	set := make(map[ModelCapability]bool, len(have))
	for _, c := range have {
		set[c] = true
	}
	for _, c := range need {
		if !set[c] {
			return false
		}
	}
	return true
}
