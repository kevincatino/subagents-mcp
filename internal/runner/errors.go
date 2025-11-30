package runner

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUsageLimitExceeded indicates the runner's API quota is exhausted.
// The caller should try the next available runner.
type ErrUsageLimitExceeded struct {
	RunnerName string
	Message    string
}

func (e *ErrUsageLimitExceeded) Error() string {
	return fmt.Sprintf("%s: usage limit exceeded: %s", e.RunnerName, e.Message)
}

// IsUsageLimitError checks if an error indicates quota exhaustion.
func IsUsageLimitError(err error) bool {
	var usageErr *ErrUsageLimitExceeded
	return errors.As(err, &usageErr)
}

// Usage limit detection patterns per runner.
var codexUsageLimitPatterns = []string{
	"you've hit your usage limit",
	"usage limit",
	"purchase more credits",
}

var copilotUsageLimitPatterns = []string{
	"usage limit",         // placeholder
	"rate limit exceeded", // placeholder
	"quota exceeded",      // placeholder
}

var geminiUsageLimitPatterns = []string{
	"usage limit",
	"quota exceeded",
	"rate limit exceeded",
	"quota has been exhausted",
}

// isCodexUsageLimitMessage checks if output contains Codex usage limit patterns.
func isCodexUsageLimitMessage(output string) bool {
	return containsUsageLimitPattern(output, codexUsageLimitPatterns)
}

// isCopilotUsageLimitMessage checks if output contains Copilot usage limit patterns.
func isCopilotUsageLimitMessage(output string) bool {
	return containsUsageLimitPattern(output, copilotUsageLimitPatterns)
}

func isGeminiUsageLimitMessage(output string) bool {
	return containsUsageLimitPattern(output, geminiUsageLimitPatterns)
}

func containsUsageLimitPattern(output string, patterns []string) bool {
	lower := strings.ToLower(output)
	for _, pattern := range patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// extractUsageLimitMessage extracts the first line containing usage limit info.
func extractUsageLimitMessage(output string) string {
	lines := strings.Split(output, "\n")
	lower := strings.ToLower(output)
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, pattern := range append(append(codexUsageLimitPatterns, copilotUsageLimitPatterns...), geminiUsageLimitPatterns...) {
			if strings.Contains(lineLower, strings.ToLower(pattern)) {
				return strings.TrimSpace(line)
			}
		}
	}
	// Fallback: return first non-empty line if pattern matched overall
	if strings.Contains(lower, "usage limit") || strings.Contains(lower, "rate limit") {
		for _, line := range lines {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				return trimmed
			}
		}
	}
	return output
}
