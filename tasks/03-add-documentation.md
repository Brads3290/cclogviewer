# Task 03: Add Concise Documentation

## Priority: 3rd (High)

## Overview
Add minimal, intent-focused documentation to exported functions and types. The code should be self-documenting; comments should explain WHY, not WHAT.

## Issues to Address
1. No package-level documentation
2. Missing godoc comments for exported types/functions
3. No intent documentation for complex algorithms

## Steps to Complete

### Step 1: Add Package Documentation
Create minimal `doc.go` files:

1. **cmd/cclogviewer/doc.go**:
```go
// Package main converts Claude Code JSONL logs to interactive HTML.
package main
```

2. **internal/models/doc.go**:
```go
// Package models defines log entry structures.
package models
```

3. Add similar one-line doc.go files for other packages

### Step 2: Document Exported Types

Keep type documentation to one line:

```go
// LogEntry represents a single JSONL log entry.
type LogEntry struct

// ProcessedEntry is a LogEntry enriched with hierarchy and metadata.
type ProcessedEntry struct
```

Only add field comments where intent is non-obvious:
```go
IsSidechain bool `json:"is_sidechain"` // Task tool conversation flag
```

### Step 3: Document Exported Functions

One-line function documentation:

```go
// ReadJSONLFile parses a JSONL file into LogEntry structs.
func ReadJSONLFile(filename string) ([]LogEntry, error)

// ProcessEntries builds a hierarchical structure from flat log entries.
func ProcessEntries(entries []models.LogEntry) []*models.ProcessedEntry
```

### Step 4: Document Algorithm Intent

Only document WHY, not HOW:

```go
// MatchToolCalls uses a 5-minute window to prevent false matches in long conversations.
func (m *ToolCallMatcher) MatchToolCalls(state *ProcessingState) error

// findBestMatchingSidechain scores matches to handle concurrent Task invocations.
func findBestMatchingSidechain(task *models.ProcessedEntry, ...) (*models.Sidechain, float64)
```

### Step 5: Skip Interface Documentation

Interfaces are self-documenting through their method signatures.

## Success Criteria
- [ ] Every package has a one-line doc.go file
- [ ] Every exported type has a one-line godoc comment
- [ ] Every exported function has a one-line description
- [ ] Complex algorithms document their intent (WHY not HOW)

## Documentation Standards
1. One complete sentence per item
2. Explain intent, not implementation
3. Let the code document itself
4. Only comment non-obvious behavior

## Notes
- Less is more
- Code clarity > documentation volume
- If you need extensive docs, refactor the code instead