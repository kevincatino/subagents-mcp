package runner

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrUsageLimitExceeded_Error(t *testing.T) {
	err := &ErrUsageLimitExceeded{
		RunnerName: "codex",
		Message:    "You've hit your usage limit",
	}

	expected := "codex: usage limit exceeded: You've hit your usage limit"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsUsageLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct usage limit error",
			err:      &ErrUsageLimitExceeded{RunnerName: "codex", Message: "limit reached"},
			expected: true,
		},
		{
			name:     "wrapped usage limit error",
			err:      fmt.Errorf("outer: %w", &ErrUsageLimitExceeded{RunnerName: "copilot", Message: "quota"}),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUsageLimitError(tt.err); got != tt.expected {
				t.Errorf("IsUsageLimitError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCodexUsageLimitMessage(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "exact match",
			output:   `{"type":"error","message":"You've hit your usage limit. Upgrade to Pro"}`,
			expected: true,
		},
		{
			name:     "case insensitive",
			output:   `YOU'VE HIT YOUR USAGE LIMIT`,
			expected: true,
		},
		{
			name:     "purchase credits pattern",
			output:   `Please purchase more credits to continue`,
			expected: true,
		},
		{
			name:     "usage limit substring",
			output:   `Error: usage limit exceeded`,
			expected: true,
		},
		{
			name:     "unrelated error",
			output:   `{"type":"error","message":"Network timeout"}`,
			expected: false,
		},
		{
			name:     "empty output",
			output:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCodexUsageLimitMessage(tt.output); got != tt.expected {
				t.Errorf("isCodexUsageLimitMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCopilotUsageLimitMessage(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "rate limit pattern",
			output:   `rate limit exceeded`,
			expected: true,
		},
		{
			name:     "quota exceeded",
			output:   `Error: quota exceeded for this account`,
			expected: true,
		},
		{
			name:     "usage limit",
			output:   `usage limit reached`,
			expected: true,
		},
		{
			name:     "unrelated error",
			output:   `{"error":"invalid token"}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCopilotUsageLimitMessage(tt.output); got != tt.expected {
				t.Errorf("isCopilotUsageLimitMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractUsageLimitMessage(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		contains string
	}{
		{
			name:     "json error message",
			output:   `{"type":"error","message":"You've hit your usage limit. Upgrade to Pro"}`,
			contains: "usage limit",
		},
		{
			name:     "multiline with pattern on second line",
			output:   "First line\nYou've hit your usage limit\nThird line",
			contains: "usage limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUsageLimitMessage(tt.output)
			if got == "" {
				t.Errorf("extractUsageLimitMessage() returned empty string")
			}
		})
	}
}
