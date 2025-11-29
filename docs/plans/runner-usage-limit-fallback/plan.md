# Runner Usage Limit Fallback Plan

## Overview

Add detection of usage/quota limit errors from runner CLIs and automatically fallback to the next available runner when the current runner is out of credits. This prevents task failures when a single runner's quota is exhausted but alternatives exist.

## Current State Analysis

- **`internal/runner/selector.go:93-110`**: `Selector.Run()` iterates candidates by model support but returns immediately on first match—no retry on failure.
- **`internal/runner/codex.go:73-77`**: Returns generic `fmt.Errorf("codex exec failed: %w; stderr: %s", ...)` without classifying errors.
- **`internal/runner/copilot.go:74-78`**: Same pattern—no error classification.
- **`internal/runner/runner.go`**: `AgentRunner` interface has only `Run()`; no runner name accessor.

### Codex Usage Limit Error (from user example)

```json
{"type":"error","message":"You've hit your usage limit. Upgrade to Pro..."}
{"type":"turn.failed","error":{"message":"You've hit your usage limit..."}}
```

Key detection string: `"You've hit your usage limit"` or `"usage limit"`.

## Desired End State

1. Runners detect usage limit errors in CLI output and return a sentinel `ErrUsageLimitExceeded` error.
2. `Selector.Run()` catches this sentinel error and continues to the next candidate runner.
3. If all runners are exhausted (either by model mismatch or usage limits), return a clear aggregated error.
4. Logging at `Warn` level when fallback occurs.

## What We're NOT Doing

- Detecting other transient errors (network, 503, timeouts) for fallback—only usage limits.
- Persisting runner state across requests (e.g., "codex is exhausted for the next hour").
- Proactive quota checking before dispatch.
- Adding retry with backoff for the same runner.

## Implementation Approach

### Error Classification

Define a sentinel error type that runners return when detecting quota exhaustion:

```go
// ErrUsageLimitExceeded indicates the runner's API quota is exhausted.
type ErrUsageLimitExceeded struct {
    RunnerName string
    Message    string
}
```

### Detection Patterns

| Runner  | Pattern                              | Source            |
|---------|--------------------------------------|-------------------|
| Codex   | `"You've hit your usage limit"`      | User example      |
| Codex   | `"usage limit"`                      | Simplified match  |
| Copilot | `"usage limit"` (placeholder)        | TBD—needs testing |
| Copilot | `"rate limit exceeded"` (placeholder)| TBD—needs testing |

### Selector Fallback Logic

Modify `Selector.Run()` to loop through all candidates, skipping to the next on `ErrUsageLimitExceeded`:

```go
for _, candidate := range candidates {
    if !supportsModel(candidate.models, model) {
        continue
    }
    output, err := candidate.runner.Run(ctx, agent, task, workdir, model)
    if err == nil {
        return output, nil
    }
    if IsUsageLimitError(err) {
        s.logger.Warn("runner hit usage limit, trying next",
            zap.String("runner", candidate.name),
            zap.Error(err))
        continue
    }
    return "", err // Non-usage-limit error: fail immediately
}
return "", fmt.Errorf("all runners exhausted or unsupported model %q", model)
```

## Phase 1: Add Sentinel Error Type

### Overview

Create a new error type for usage limit detection and a helper to check for it.

### Changes Required

#### 1. New Error Type

**File**: `internal/runner/errors.go` (new)

```go
package runner

import (
    "errors"
    "fmt"
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
```

### Tests

**File**: `internal/runner/errors_test.go` (new)

- `TestErrUsageLimitExceeded_Error` — verify message format.
- `TestIsUsageLimitError` — verify `errors.As` detection works, including wrapped errors.

---

## Phase 2: Detect Usage Limit in Codex Runner

### Overview

Parse Codex CLI stderr/stdout for usage limit patterns and return the sentinel error.

### Changes Required

#### 1. Add Detection Logic

**File**: `internal/runner/codex.go`

After `cmd.Run()`, before returning the generic error, check if stderr or stdout contains usage limit patterns:

```go
if err != nil {
    combined := stderr.String() + stdout.String()
    if isUsageLimitMessage(combined) {
        return "", &ErrUsageLimitExceeded{
            RunnerName: "codex",
            Message:    extractUsageLimitMessage(combined),
        }
    }
    return "", fmt.Errorf("codex exec failed: %w; stderr: %s", err, stderr.String())
}
```

#### 2. Pattern Matching Helper

**File**: `internal/runner/codex.go` (or shared in `errors.go`)

```go
var codexUsageLimitPatterns = []string{
    "You've hit your usage limit",
    "usage limit",
    "purchase more credits",
}

func isUsageLimitMessage(output string) bool {
    lower := strings.ToLower(output)
    for _, pattern := range codexUsageLimitPatterns {
        if strings.Contains(lower, strings.ToLower(pattern)) {
            return true
        }
    }
    return false
}
```

### Tests

**File**: `internal/runner/codex_test.go`

- `TestCodexRunner_UsageLimitDetection` — mock CLI returning usage limit JSON, verify `ErrUsageLimitExceeded` is returned.
- `TestCodexRunner_OtherErrorNotUsageLimit` — mock CLI returning different error, verify generic error.

---

## Phase 3: Detect Usage Limit in Copilot Runner

### Overview

Add placeholder patterns for Copilot usage limit detection.

### Changes Required

#### 1. Add Detection Logic

**File**: `internal/runner/copilot.go`

Same pattern as Codex:

```go
if err != nil {
    combined := stderr.String() + stdout.String()
    if isCopilotUsageLimitMessage(combined) {
        return "", &ErrUsageLimitExceeded{
            RunnerName: "copilot",
            Message:    extractUsageLimitMessage(combined),
        }
    }
    return "", fmt.Errorf("copilot exec failed: %w; stderr: %s", err, stderr.String())
}
```

#### 2. Placeholder Patterns

```go
var copilotUsageLimitPatterns = []string{
    "usage limit",           // placeholder
    "rate limit exceeded",   // placeholder
    "quota exceeded",        // placeholder
}
```

### Tests

**File**: `internal/runner/copilot_test.go`

- `TestCopilotRunner_UsageLimitDetection` — mock CLI returning usage limit output.

---

## Phase 4: Update Selector for Fallback

### Overview

Modify `Selector.Run()` to catch usage limit errors and try the next runner.

### Changes Required

#### 1. Add Logger to Selector

**File**: `internal/runner/selector.go`

The `Selector` struct needs a logger field to emit warnings:

```go
type Selector struct {
    logger    *zap.Logger
    preferred *namedRunner
    fallbacks []namedRunner
}
```

Update `NewSelector` to store the logger.

#### 2. Implement Fallback Loop

**File**: `internal/runner/selector.go`

Replace the current `Run()` implementation:

```go
func (s *Selector) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
    candidates := make([]namedRunner, 0, 1+len(s.fallbacks))
    if s.preferred != nil {
        candidates = append(candidates, *s.preferred)
    }
    candidates = append(candidates, s.fallbacks...)

    var lastUsageLimitErr error
    for _, candidate := range candidates {
        if !supportsModel(candidate.models, model) {
            continue
        }
        output, err := candidate.runner.Run(ctx, agent, task, workdir, model)
        if err == nil {
            return output, nil
        }
        if IsUsageLimitError(err) {
            s.logger.Warn("runner hit usage limit, trying next",
                zap.String("runner", candidate.name),
                zap.Error(err))
            lastUsageLimitErr = err
            continue
        }
        // Non-usage-limit error: fail immediately
        return "", err
    }

    if lastUsageLimitErr != nil {
        return "", fmt.Errorf("all runners exhausted due to usage limits: %w", lastUsageLimitErr)
    }
    if model == "" {
        return "", fmt.Errorf("no runner available")
    }
    return "", fmt.Errorf("no runner supports model %q", model)
}
```

### Tests

**File**: `internal/runner/selector_test.go`

- `TestSelector_FallbackOnUsageLimit` — first runner returns `ErrUsageLimitExceeded`, second succeeds.
- `TestSelector_AllRunnersExhausted` — all runners return usage limit errors.
- `TestSelector_NonUsageLimitErrorNoFallback` — first runner returns generic error, no fallback attempted.

---

## Phase 5: Integration & Documentation

### Overview

Wire everything together and update docs.

### Changes Required

#### 1. Verify Handler Integration

**File**: `internal/mcp/handlers.go`

No changes needed—`Handlers` calls `runner.Run()` which is now the `Selector`, and fallback is transparent.

#### 2. Update Documentation

**File**: `docs/architecture.md`

Add a section on runner fallback behavior:

> When a runner returns a usage limit error (e.g., Codex quota exhausted), the selector automatically tries the next runner by priority. Non-usage-limit errors fail immediately.

**File**: `docs/setup.md`

Note that multiple runners can be configured for redundancy.

---

## Work Plan

| Phase | Task | Files | Estimated Effort |
|-------|------|-------|------------------|
| 1 | Add `ErrUsageLimitExceeded` and `IsUsageLimitError` | `internal/runner/errors.go`, `errors_test.go` | Small |
| 2 | Detect usage limit in Codex runner | `internal/runner/codex.go`, `codex_test.go` | Small |
| 3 | Detect usage limit in Copilot runner (placeholder) | `internal/runner/copilot.go`, `copilot_test.go` | Small |
| 4 | Update Selector with fallback loop + logger | `internal/runner/selector.go`, `selector_test.go` | Medium |
| 5 | Documentation updates | `docs/architecture.md`, `docs/setup.md` | Small |

## Success Criteria

1. When Codex returns `"You've hit your usage limit"`, the task automatically retries with Copilot (if configured and model-compatible).
2. When all runners are exhausted, a clear error is returned listing the cause.
3. Non-usage-limit errors fail fast without fallback attempts.
4. Fallback events are logged at `Warn` level with runner name and error.
5. All new code has unit test coverage.

## Open Issues

None currently. Copilot patterns are placeholders pending real-world testing.
