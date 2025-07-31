# Task 06: Remove Dead Code

## Priority: 6th (Medium)

## Overview
The codebase contains approximately 25+ dead code items including unused functions, an entire unused processing architecture, unused struct fields, and duplicate implementations. This dead code creates confusion and increases maintenance burden.

## Issues to Address
1. Entire unused Chain of Responsibility architecture
2. Unused functions and methods (11+ items)
3. Unused struct fields (5+ items)
4. Duplicate helper functions
5. Unused variables

## Steps to Complete

### Step 1: Remove Unused Processing Architecture

**Location**: `internal/processor/processor.go`

Delete the following unused types and methods:
```go
// DELETE: Lines 9-13 - EntryProcessor interface
type EntryProcessor interface {
    CanProcess(entry *models.LogEntry) bool
    Process(entry *models.LogEntry, state *ProcessingState) error
}

// DELETE: Lines 15-18 - ProcessorChain struct
type ProcessorChain struct {
    processors []EntryProcessor
}

// DELETE: Lines 20-29 - NewProcessorChain function
func NewProcessorChain() *ProcessorChain { ... }

// DELETE: Lines 31-39 - Process method
func (pc *ProcessorChain) Process(entries []models.LogEntry) error { ... }

// DELETE: Lines 41-64 - ToolCallProcessor struct and methods
type ToolCallProcessor struct{}
func (tcp *ToolCallProcessor) CanProcess(...) bool { ... }
func (tcp *ToolCallProcessor) Process(...) error { ... }

// DELETE: Lines 66-80 - ToolResultProcessor struct and methods
type ToolResultProcessor struct{}
func (trp *ToolResultProcessor) CanProcess(...) bool { ... }
func (trp *ToolResultProcessor) Process(...) error { ... }

// DELETE: Lines 82-93 - MessageProcessor struct and methods  
type MessageProcessor struct{}
func (mp *MessageProcessor) CanProcess(...) bool { ... }
func (mp *MessageProcessor) Process(...) error { ... }
```

After deletion, the file should only contain the `parseTimestamp` function (if it's used elsewhere).

### Step 2: Remove Unused Hierarchy Functions

**Location**: `internal/processor/hierarchy.go`

Delete the following unused methods:
```go
// DELETE: Lines 52-56 - calculateDepths (comment says "now handled by setEntryDepth")
func (h *HierarchyBuilder) calculateDepths() { ... }

// DELETE: Lines 58-67 - findRootEntries
func (h *HierarchyBuilder) findRootEntries() []*models.ProcessedEntry { ... }

// DELETE: Lines 69-80 - BuildParentChildMap
func (h *HierarchyBuilder) BuildParentChildMap() map[string][]*models.ProcessedEntry { ... }
```

### Step 3: Remove Unused Sidechain Functions

**Location**: `internal/processor/sidechain.go`

Delete the following unused methods:
```go
// DELETE: Lines 285-292 - identifySidechainBoundaries
func (s *SidechainProcessor) identifySidechainBoundaries() { ... }

// DELETE: Lines 295-300 - groupSidechainEntries  
func (s *SidechainProcessor) groupSidechainEntries() map[string][]*models.ProcessedEntry { ... }
```

### Step 4: Remove Unused Matcher Functions

**Location**: `internal/processor/matcher.go`

Delete the following unused methods and fields:
```go
// DELETE: Line 11 - unused windowSize field
windowSize time.Duration

// DELETE: Lines 88-91 - findToolCall
func (m *ToolCallMatcher) findToolCall(uuid string) *models.ProcessedEntry { ... }

// DELETE: Lines 93-100 - isWithinWindow
func (m *ToolCallMatcher) isWithinWindow(t1, t2 time.Time) bool { ... }
```

Update the `NewToolCallMatcher` constructor to remove windowSize parameter.

### Step 5: Remove Duplicate Functions

**Location**: `internal/processor/entries.go`

Delete the following duplicate/unused functions:
```go
// DELETE: Lines 281-287 - truncateString (unused)
func truncateString(s string, maxLen int) string { ... }

// DELETE: Lines 289-328 - extractFullSidechainContent (unused)
func extractFullSidechainContent(...) string { ... }

// DELETE: Lines 412-429 - extractTaskPrompt (duplicate exists in sidechain.go)
func extractTaskPrompt(toolCall *models.ProcessedEntry) string { ... }

// DELETE: Lines 431-441 - normalizeText (if duplicate exists elsewhere)
func normalizeText(text string) string { ... }

// DELETE: Lines 479-505 - setEntryDepth (duplicate exists in hierarchy.go)
func setEntryDepth(entry *models.ProcessedEntry, depth int) { ... }
```

### Step 6: Remove Unused Struct Fields

**Location**: `internal/processor/models.go`

Delete or mark as deprecated unused fields:
```go
type ProcessingState struct {
    Entries      []*models.ProcessedEntry
    EntryMap     map[string]*models.ProcessedEntry
    // DELETE: RootParent field (line 13) - never used
    // RootParent   *models.ProcessedEntry
    ToolCallMap  map[string]*ToolCallContext
    Sidechains   []*SidechainContext
    ParentChildMap map[string][]*models.ProcessedEntry
}

type ToolCallContext struct {
    ToolCall  *models.ProcessedEntry
    Result    *models.ProcessedEntry
    // DELETE: ParentID field (line 22) - never used
    // ParentID  string
    // DELETE: IsComplete field (line 23) - never used  
    // IsComplete bool
}

type SidechainContext struct {
    RootEntry *models.ProcessedEntry
    Entries   []*models.ProcessedEntry
    // DELETE: StartIndex field (line 29) - never used
    // StartIndex int
    // DELETE: EndIndex field (line 30) - never used
    // EndIndex   int
}

// DELETE: Lines 34-37 - MatchingOptions struct (entire struct unused)
// type MatchingOptions struct { ... }
```

### Step 7: Remove Unused Tool Functions

**Location**: `internal/processor/tools.go`

Delete the following:
```go
// DELETE: Lines 54-61 - formatBashToolInput
// Comment says "kept for backward compatibility" but not used
func formatBashToolInput(input map[string]interface{}, cwd string) template.HTML { ... }
```

### Step 8: Remove Unused Formatter Methods

**Location**: `internal/processor/tools/formatters/base.go`

Check if `FormatOutput` is used anywhere:
```go
// If unused, DELETE: Lines 19-23
func (b *BaseFormatter) FormatOutput(data map[string]interface{}) (template.HTML, error) {
    return "", fmt.Errorf("FormatOutput not implemented")
}
```

### Step 9: Clean Up Imports

After removing dead code, clean up unused imports in each file:
1. Run `goimports -w .` to remove unused imports
2. Or manually remove imports that are no longer needed

### Step 10: Verify No Broken References

1. Run build to ensure no broken references:
```bash
go build ./...
```

2. Run tests to ensure nothing broke:
```bash
go test ./...
```

3. Use static analysis to find any remaining dead code:
```bash
# Install deadcode tool
go install golang.org/x/tools/cmd/deadcode@latest

# Run deadcode analysis
deadcode ./...
```

## Success Criteria
- [ ] All identified dead code removed
- [ ] No broken references after removal
- [ ] All tests still pass
- [ ] Build succeeds without errors
- [ ] Code coverage doesn't decrease
- [ ] Static analysis shows no dead code

## Safety Checklist

Before removing each piece of code:
1. Search entire codebase for references
2. Check if it's used in templates
3. Verify it's not used via reflection
4. Check if it's part of a public API
5. Look for indirect usage through interfaces

## Tools to Use
- `grep -r "functionName" .` - Find references
- `go build ./...` - Verify build
- `go test ./...` - Run tests
- `deadcode` - Find dead code
- `unused` - Another dead code detector

## Post-Cleanup Tasks
1. Update documentation to reflect removed code
2. Update any architecture diagrams
3. Commit with clear message about what was removed
4. Consider adding linter rules to prevent future dead code

## Notes
- Be conservative - if unsure whether code is used, investigate thoroughly
- Some "dead" code might be there for future use - check with team
- Keep a list of what was removed in case rollback is needed
- Consider deprecating before removing if there's any doubt