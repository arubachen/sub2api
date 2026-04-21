package service

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	openaipkg "github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

func NormalizeOpenAICompatRequestedModel(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}
	trimmed = stripPresentedOpenAICompatModelPrefix(trimmed)

	normalized, _, ok := splitOpenAICompatReasoningModel(trimmed)
	if !ok || normalized == "" {
		return trimmed
	}
	return normalized
}

func BuildOpenAICompatModelCatalogEntry(model string) openaipkg.Model {
	rawModel := strings.TrimSpace(model)
	owner := inferOpenAICompatModelOwner(rawModel)

	return openaipkg.Model{
		ID:          rawModel,
		Name:        rawModel,
		Object:      "model",
		Created:     1704067200, // 2024-01-01T00:00:00Z
		OwnedBy:     owner,
		Type:        "model",
		DisplayName: rawModel,
	}
}

func stripPresentedOpenAICompatModelPrefix(model string) string {
	trimmed := strings.TrimSpace(model)
	prefix, remainder, ok := strings.Cut(trimmed, "/")
	if !ok {
		return trimmed
	}

	prefix = canonicalOpenAICompatModelOwner(prefix)
	remainder = strings.TrimSpace(remainder)
	if prefix == "" || remainder == "" {
		return trimmed
	}

	if prefix == canonicalOpenAICompatModelOwner(inferOpenAICompatModelOwner(remainder)) {
		return remainder
	}
	return trimmed
}

func inferOpenAICompatModelOwner(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}
	if prefix, _, ok := strings.Cut(trimmed, "/"); ok {
		return canonicalOpenAICompatModelOwner(prefix)
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "gpt-"),
		strings.HasPrefix(lower, "o1"),
		strings.HasPrefix(lower, "o3"),
		strings.HasPrefix(lower, "o4"),
		strings.HasPrefix(lower, "text-embedding-"),
		strings.HasPrefix(lower, "omni-moderation"):
		return "openai"
	case strings.HasPrefix(lower, "grok"):
		return "x-ai"
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	case strings.HasPrefix(lower, "gemini"):
		return "google"
	case strings.HasPrefix(lower, "qwen"):
		return "qwen"
	case strings.HasPrefix(lower, "deepseek"):
		return "deepseek"
	case strings.HasPrefix(lower, "kimi"), strings.HasPrefix(lower, "moonshot"):
		return "moonshot"
	case strings.HasPrefix(lower, "minimax"):
		return "minimax"
	case strings.HasPrefix(lower, "mistral"):
		return "mistral"
	case strings.HasPrefix(lower, "glm"):
		return "z-ai"
	}

	first := strings.FieldsFunc(lower, func(r rune) bool {
		switch r {
		case '-', '_', ' ', ':':
			return true
		default:
			return false
		}
	})
	if len(first) == 0 {
		return ""
	}
	return canonicalOpenAICompatModelOwner(first[0])
}

func canonicalOpenAICompatModelOwner(owner string) string {
	switch strings.ToLower(strings.TrimSpace(owner)) {
	case "openai", "gpt", "oai":
		return "openai"
	case "xai", "x-ai", "grok":
		return "x-ai"
	case "google", "googleai", "gemini":
		return "google"
	case "anthropic", "claude":
		return "anthropic"
	case "zai", "z-ai", "zhipu", "glm":
		return "z-ai"
	case "moonshot", "kimi":
		return "moonshot"
	case "mistralai", "mistral":
		return "mistral"
	default:
		return strings.ToLower(strings.TrimSpace(owner))
	}
}

func applyOpenAICompatModelNormalization(req *apicompat.AnthropicRequest) {
	if req == nil {
		return
	}

	originalModel := strings.TrimSpace(req.Model)
	if originalModel == "" {
		return
	}

	normalizedModel, derivedEffort, hasReasoningSuffix := splitOpenAICompatReasoningModel(originalModel)
	if hasReasoningSuffix && normalizedModel != "" {
		req.Model = normalizedModel
	}

	if req.OutputConfig != nil && strings.TrimSpace(req.OutputConfig.Effort) != "" {
		return
	}

	claudeEffort := openAIReasoningEffortToClaudeOutputEffort(derivedEffort)
	if claudeEffort == "" {
		return
	}

	if req.OutputConfig == nil {
		req.OutputConfig = &apicompat.AnthropicOutputConfig{}
	}
	req.OutputConfig.Effort = claudeEffort
}

func splitOpenAICompatReasoningModel(model string) (normalizedModel string, reasoningEffort string, ok bool) {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return "", "", false
	}

	modelID := trimmed
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
	}
	modelID = strings.TrimSpace(modelID)
	if !strings.HasPrefix(strings.ToLower(modelID), "gpt-") {
		return trimmed, "", false
	}

	parts := strings.FieldsFunc(strings.ToLower(modelID), func(r rune) bool {
		switch r {
		case '-', '_', ' ':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return trimmed, "", false
	}

	last := strings.NewReplacer("-", "", "_", "", " ", "").Replace(parts[len(parts)-1])
	switch last {
	case "none", "minimal":
	case "low", "medium", "high":
		reasoningEffort = last
	case "xhigh", "extrahigh":
		reasoningEffort = "xhigh"
	default:
		return trimmed, "", false
	}

	return normalizeCodexModel(modelID), reasoningEffort, true
}

func openAIReasoningEffortToClaudeOutputEffort(effort string) string {
	switch strings.TrimSpace(effort) {
	case "low", "medium", "high":
		return effort
	case "xhigh":
		return "max"
	default:
		return ""
	}
}
