# Task 02: Refactor entries.go

**File:** `internal/processor/entries.go`  
**Current Size:** 752 lines  
**Priority:** HIGH  
**Estimated Effort:** 6-8 hours  

## Problem Summary

The `entries.go` file contains a monolithic `ProcessEntries()` function that spans 300+ lines (lines 229-531) and handles multiple complex responsibilities:
- Tool call matching
- Sidechain conversation grouping
- Entry hierarchy building
- Depth calculation
- State management

The function has deep nesting (up to 5 levels), making it difficult to understand, test, and maintain.

## Current State Analysis

### Main Issues

1. **God Function:** `ProcessEntries()` does everything:
   - Iterates through entries
   - Matches tool calls with results
   - Handles sidechain (Task tool) conversations
   - Builds parent-child relationships
   - Calculates display depths
   - Manages complex state

2. **Deep Nesting:** Multiple levels of conditions:
   ```go
   for _, entry := range entries {
       if entry.Type == "tool_call" {
           if toolCall := entry.ToolCall; toolCall != nil {
               if toolCall.Name == "Task" {
                   // More nesting...
               }
           }
       }
   }
   ```

3. **Complex State Management:**
   - `toolCallMap` for tracking pending tool calls
   - `parentChildMap` for hierarchy
   - `rootParent` for sidechain relationships
   - Multiple interdependent data structures

4. **Mixed Concerns:**
   - Business logic mixed with data structure manipulation
   - Tool-specific logic embedded in general processing

## Proposed Solution

### Architecture Redesign

```
processor/
├── entries.go          # Entry point and orchestration
├── matcher.go          # Tool call matching logic
├── hierarchy.go        # Parent-child relationship building
├── sidechain.go        # Sidechain conversation handling
├── state.go            # Processing state management
└── models.go           # Internal data structures
```

### Key Abstractions

1. **ProcessingState struct:**
   ```go
   type ProcessingState struct {
       Entries          []*ProcessedEntry
       ToolCallMap      map[string]*ToolCallContext
       ParentChildMap   map[string][]string
       CurrentSidechain *SidechainContext
   }
   ```

2. **EntryProcessor interface:**
   ```go
   type EntryProcessor interface {
       Process(entry *models.LogEntry, state *ProcessingState) error
   }
   ```

3. **Specialized processors:**
   - `ToolCallProcessor`
   - `ToolResultProcessor`
   - `SidechainProcessor`
   - `MessageProcessor`

## Implementation Steps

### Step 1: Extract Data Structures (1 hour)

Create `internal/processor/models.go`:

```go
package processor

// ProcessingState holds all state during entry processing
type ProcessingState struct {
    Entries        []*ProcessedEntry
    ToolCallMap    map[string]*ToolCallContext
    ParentChildMap map[string][]string
    RootParent     string
    Index          int
}

// ToolCallContext tracks pending tool calls
type ToolCallContext struct {
    Entry      *ProcessedEntry
    CallTime   time.Time
    ParentID   string
    IsComplete bool
}

// SidechainContext tracks active sidechain conversations
type SidechainContext struct {
    RootToolCallID string
    StartIndex     int
    EndIndex       int
    Entries        []*ProcessedEntry
}
```

### Step 2: Create Entry Processor Interface (30 min)

Create `internal/processor/processor.go`:

```go
package processor

type EntryProcessor interface {
    CanProcess(entry *models.LogEntry) bool
    Process(entry *models.LogEntry, state *ProcessingState) (*ProcessedEntry, error)
}

type ProcessorChain struct {
    processors []EntryProcessor
}

func (pc *ProcessorChain) Process(entry *models.LogEntry, state *ProcessingState) (*ProcessedEntry, error) {
    for _, processor := range pc.processors {
        if processor.CanProcess(entry) {
            return processor.Process(entry, state)
        }
    }
    return nil, fmt.Errorf("no processor found for entry type: %s", entry.Type)
}
```

### Step 3: Extract Tool Call Matching (2 hours)

Create `internal/processor/matcher.go`:

```go
package processor

type ToolCallMatcher struct {
    windowSize time.Duration
}

func (m *ToolCallMatcher) MatchToolCalls(entries []*ProcessedEntry) error {
    // Extract tool call matching logic from ProcessEntries
    // This includes:
    // - Building the tool call map
    // - Matching results to calls
    // - Handling edge cases (missing results, etc.)
}

func (m *ToolCallMatcher) findToolCall(state *ProcessingState, toolUseID string) *ToolCallContext {
    // Extract logic for finding matching tool calls
}

func (m *ToolCallMatcher) isWithinWindow(callTime, resultTime time.Time) bool {
    // Time window validation logic
}
```

### Step 4: Extract Hierarchy Building (1.5 hours)

Create `internal/processor/hierarchy.go`:

```go
package processor

type HierarchyBuilder struct{}

func (h *HierarchyBuilder) BuildHierarchy(entries []*ProcessedEntry) error {
    // Extract parent-child relationship logic
    // Calculate depths
    // Handle root entries
}

func (h *HierarchyBuilder) calculateDepths(entries []*ProcessedEntry, parentChildMap map[string][]string) {
    // Extract depth calculation algorithm
}

func (h *HierarchyBuilder) findRootEntries(entries []*ProcessedEntry) []*ProcessedEntry {
    // Logic to identify root-level entries
}
```

### Step 5: Extract Sidechain Processing (2 hours)

Create `internal/processor/sidechain.go`:

```go
package processor

type SidechainProcessor struct{}

func (s *SidechainProcessor) ProcessSidechains(entries []*ProcessedEntry) error {
    // Extract all sidechain-related logic
    // Group sidechain conversations
    // Link to parent tool calls
}

func (s *SidechainProcessor) identifySidechainBoundaries(entries []*ProcessedEntry) []SidechainContext {
    // Logic to find sidechain start/end points
}

func (s *SidechainProcessor) groupSidechainEntries(ctx SidechainContext, entries []*ProcessedEntry) *ProcessedEntry {
    // Create grouped sidechain entry
}
```

### Step 6: Refactor Main ProcessEntries (1.5 hours)

Refactor `ProcessEntries()` to use the new components:

```go
func ProcessEntries(entries []*models.LogEntry) ([]*ProcessedEntry, error) {
    // Initialize state
    state := &ProcessingState{
        Entries:        make([]*ProcessedEntry, 0, len(entries)),
        ToolCallMap:    make(map[string]*ToolCallContext),
        ParentChildMap: make(map[string][]string),
    }
    
    // Phase 1: Convert entries using processor chain
    chain := NewProcessorChain()
    for i, entry := range entries {
        state.Index = i
        processed, err := chain.Process(entry, state)
        if err != nil {
            return nil, fmt.Errorf("processing entry %d: %w", i, err)
        }
        state.Entries = append(state.Entries, processed)
    }
    
    // Phase 2: Match tool calls
    matcher := NewToolCallMatcher()
    if err := matcher.MatchToolCalls(state.Entries); err != nil {
        return nil, fmt.Errorf("matching tool calls: %w", err)
    }
    
    // Phase 3: Process sidechains
    sidechainProc := NewSidechainProcessor()
    if err := sidechainProc.ProcessSidechains(state.Entries); err != nil {
        return nil, fmt.Errorf("processing sidechains: %w", err)
    }
    
    // Phase 4: Build hierarchy
    hierarchy := NewHierarchyBuilder()
    if err := hierarchy.BuildHierarchy(state.Entries); err != nil {
        return nil, fmt.Errorf("building hierarchy: %w", err)
    }
    
    return filterVisibleEntries(state.Entries), nil
}
```

### Step 7: Add Unit Tests (1.5 hours)

Create comprehensive tests for each component:
- `matcher_test.go` - Test tool call matching edge cases
- `hierarchy_test.go` - Test depth calculation and parent-child relationships
- `sidechain_test.go` - Test sidechain grouping logic
- `processor_test.go` - Test individual processors

## Benefits

1. **Maintainability:** Each component has a single responsibility
2. **Testability:** Small, focused functions are easier to test
3. **Readability:** Clear separation of concerns, reduced nesting
4. **Extensibility:** Easy to add new entry types or processing logic
5. **Debugging:** Clearer flow, easier to trace issues

## Risks and Mitigation

1. **Risk:** Breaking existing functionality during refactoring
   - **Mitigation:** Comprehensive test suite before refactoring
   - **Mitigation:** Refactor in small, testable increments

2. **Risk:** Performance regression from additional abstraction
   - **Mitigation:** Benchmark critical paths
   - **Mitigation:** Keep hot paths optimized

3. **Risk:** Over-engineering with too many abstractions
   - **Mitigation:** Only extract clear, cohesive responsibilities
   - **Mitigation:** Keep interfaces simple

## Success Criteria

- [ ] `ProcessEntries()` reduced to <50 lines
- [ ] No functions exceed 50 lines
- [ ] Maximum nesting depth of 3 levels
- [ ] 80%+ test coverage for new components
- [ ] No functional regressions
- [ ] Performance within 5% of original

## Future Enhancements

This refactoring enables:
- Plugin system for custom entry processors
- Parallel processing of independent entries
- Streaming processing for large files
- Custom tool call matching strategies
- Configurable processing pipelines