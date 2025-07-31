# Task 01: Refactor Architecture

## Priority: 1st (Critical)

## Overview
The current architecture has a 506-line god function (`ProcessEntries`), a 27-field god struct (`ProcessedEntry`), and abandoned code. This task will incrementally refactor the code to be more maintainable without a complete architectural overhaul.

## Issues to Address
1. `ProcessEntries()` function doing 7 different things (God Function - 506 lines)
2. `ProcessedEntry` struct with 27 fields mixing multiple concerns (God Class)
3. Tight coupling within processor package
4. Abandoned Chain of Responsibility pattern with unused code
5. Scattered tool processing logic

## Steps to Complete

### Step 1: Remove Dead Code (Quick Win)
1. Delete `internal/processor/processor.go` (entire file - unused Chain of Responsibility)
2. Remove duplicate functions:
   - Duplicate `setEntryDepth` implementations
   - Duplicate `extractTaskPrompt` functions
3. Remove unused struct fields and methods
4. Clean up unused imports

### Step 2: Extract Processing Phases from ProcessEntries
Break the 506-line function into smaller, focused functions while keeping the same structure:

```go
func ProcessEntries(entries []models.LogEntry) []*models.ProcessedEntry {
    state := initializeProcessingState(len(entries))
    entryMap := make(map[string]*models.ProcessedEntry)
    
    // Phase 1: Process all entries
    processAllEntries(entries, state, entryMap)
    
    // Phase 2: Match tool calls with results  
    matchToolCallsWithResults(state)
    
    // Phase 3: Process sidechains
    processSidechainConversations(state, entries, entryMap)
    
    // Phase 4-7: Post-processing
    rootEntries := getRootEntries(state)
    calculateAllTokens(rootEntries)
    checkAllMissingResults(rootEntries)
    linkAllCommandOutputs(rootEntries)
    buildFinalHierarchy(rootEntries)
    
    return rootEntries
}
```

Create `internal/processor/phases.go` to hold the extracted phase functions.

### Step 3: Group ProcessedEntry Fields
Keep ProcessedEntry but organize related fields using embedded structs:

```go
// TokenMetrics groups token-related fields
type TokenMetrics struct {
    TokenCount          int
    TotalTokens         int
    InputTokens         int
    OutputTokens        int
    CacheReadTokens     int
    CacheCreationTokens int
}

// CommandInfo groups command-related fields
type CommandInfo struct {
    IsCommandMessage bool
    CommandName      string
    CommandArgs      string
    CommandOutput    string
}

// ProcessedEntry becomes cleaner
type ProcessedEntry struct {
    // Core fields
    UUID         string
    ParentUUID   string
    Type         string
    Timestamp    string
    RawTimestamp string
    Role         string
    Content      string
    
    // Relationships
    Children     []*ProcessedEntry
    Depth        int
    
    // Tool-related
    ToolCalls    []ToolCall
    IsToolResult bool
    ToolResultID string
    
    // Embedded structs for grouping
    TokenMetrics
    CommandInfo
    
    // Flags
    IsSidechain     bool
    IsError         bool
    IsCaveatMessage bool
}
```

### Step 4: Consolidate Tool Processing
Create `internal/processor/tool_processor.go` to centralize scattered tool logic:

```go
type ToolProcessor struct {
    formatters map[string]ToolFormatter
}

func NewToolProcessor() *ToolProcessor {
    return &ToolProcessor{
        formatters: initializeFormatters(),
    }
}

func (tp *ToolProcessor) ProcessToolCall(toolCall *ToolCall) {
    if formatter, ok := tp.formatters[toolCall.Name]; ok {
        toolCall.FormattedContent = formatter.Format(toolCall)
    }
}
```

### Step 5: Extract Constants
Create `internal/processor/constants.go`:

```go
const (
    // Tool names
    ToolNameTask = "Task"
    ToolNameBash = "Bash"
    ToolNameEdit = "Edit"
    
    // Entry types
    TypeMessage = "message"
    TypeToolUse = "tool_use"
    TypeToolResult = "tool_result"
    
    // Roles
    RoleUser = "user"
    RoleAssistant = "assistant"
)
```

### Step 6: Simplify State Management
Instead of passing around the entire ProcessingState, use focused parameters:

```go
// Before: Everything uses ProcessingState
func MatchToolCalls(state *ProcessingState) error

// After: Use only what's needed
func MatchToolCalls(entries []*ProcessedEntry, toolCallMap map[string]*ToolCallContext) error
```

### Step 7: Add Focused Unit Tests
As each function is extracted, add tests:

```go
func TestProcessAllEntries(t *testing.T) {
    // Test the extracted phase function
}

func TestCalculateTokenMetrics(t *testing.T) {
    // Test token calculation in isolation
}
```

## Testing Strategy
1. Write unit tests for each extracted function
2. Use existing test files as regression tests
3. Add tests incrementally as code is refactored

## Success Criteria
- [ ] ProcessEntries reduced from 506 lines to < 50 lines
- [ ] Each processing phase in its own testable function
- [ ] ProcessedEntry fields grouped logically
- [ ] All dead code removed (processor.go, duplicates)
- [ ] Tool processing consolidated in one place
- [ ] Constants extracted from magic strings
- [ ] Tests added for new functions

## Potential Risks
- Breaking existing functionality during refactoring
- Need to maintain backward compatibility with existing HTML output
- Performance impact from additional function calls (minimal)

## Notes
- Keep changes incremental and testable
- Each step should leave the code working
- Prioritize readability and maintainability over perfect architecture
- This sets the foundation for future improvements without over-engineering