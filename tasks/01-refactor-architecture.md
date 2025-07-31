# Task 01: Refactor Architecture

## Priority: 1st (Critical)

## Overview
The current architecture has significant issues including god objects, circular dependencies, and poor separation of concerns. This task will restructure the code to follow clean architecture principles.

## Issues to Address
1. `ProcessEntries()` function doing 7 different things (God Function)
2. `ProcessedEntry` struct with 25+ fields mixing multiple concerns (God Class)
3. Circular dependencies throughout processor package
4. Abandoned Chain of Responsibility pattern with unused code
5. Poor separation between domain logic and presentation

## Steps to Complete

### Step 1: Analyze Current Dependencies
1. Map out all current dependencies between packages
2. Identify circular dependencies
3. Document the current processing flow

### Step 2: Create New Package Structure
```
internal/
├── domain/           # Pure business logic, no external dependencies
│   ├── models/      # Core domain models
│   └── services/    # Domain services
├── application/     # Application services, use cases
│   ├── processors/  # Processing pipeline components
│   └── mappers/     # Domain to DTO mappers
├── infrastructure/  # External concerns
│   ├── parser/      # JSONL parsing
│   ├── renderer/    # HTML rendering
│   └── browser/     # Browser launching
└── interfaces/      # Interfaces and DTOs
    ├── dto/         # Data transfer objects
    └── ports/       # Interface definitions
```

### Step 3: Split Domain Models
1. Create pure domain model:
```go
// domain/models/log_entry.go
type LogEntry struct {
    UUID      string
    Type      string
    Timestamp time.Time
    Message   json.RawMessage
}
```

2. Create view model for rendering:
```go
// interfaces/dto/log_entry_view.go
type LogEntryView struct {
    UUID         string
    DisplayDepth int
    Content      template.HTML
    // Only presentation concerns
}
```

3. Create processing model:
```go
// application/models/processing_entry.go
type ProcessingEntry struct {
    Entry    *domain.LogEntry
    Metadata ProcessingMetadata
    // Processing state
}
```

### Step 4: Refactor ProcessEntries into Pipeline
1. Create pipeline interface:
```go
// application/processors/pipeline.go
type Pipeline interface {
    Process(entries []domain.LogEntry) ([]dto.LogEntryView, error)
}

type Stage interface {
    Process(ctx *ProcessingContext) error
    Name() string
}
```

2. Break ProcessEntries into stages:
- ParseStage: Initial parsing and validation
- ToolMatchingStage: Match tool calls with results
- SidechainStage: Process sidechain conversations
- TokenCalculationStage: Calculate token usage
- HierarchyStage: Build parent-child relationships
- RenderingStage: Convert to view models

3. Implement each stage as a separate, testable component

### Step 5: Remove Circular Dependencies
1. Extract shared utilities to `internal/utils` package
2. Use dependency injection for cross-cutting concerns
3. Define clear interfaces between layers
4. Ensure dependencies only flow inward (infrastructure → application → domain)

### Step 6: Clean Up Dead Code
1. Delete unused processor implementations:
   - `ToolCallProcessor`
   - `ToolResultProcessor` 
   - `MessageProcessor`
   - `ProcessorChain`
2. Remove duplicate functions:
   - Duplicate `setEntryDepth`
   - Duplicate `extractTaskPrompt`
3. Remove unused struct fields and methods

### Step 7: Implement Dependency Injection
1. Create a container or factory for wiring dependencies
2. Pass dependencies explicitly rather than using package-level variables
3. Remove global state (e.g., `debug.Enabled`)

### Step 8: Add Integration Points
1. Define clear interfaces for each component
2. Ensure components are independently testable
3. Add factory methods for creating configured instances

## Testing Strategy
1. Write unit tests for each pipeline stage
2. Create integration tests for the full pipeline
3. Ensure each component can be tested in isolation

## Success Criteria
- [ ] No circular dependencies between packages
- [ ] Each function has a single responsibility
- [ ] Clear separation between domain, application, and infrastructure
- [ ] All dead code removed
- [ ] Each component is independently testable
- [ ] Processing pipeline is modular and extensible

## Potential Risks
- Large refactoring may introduce bugs
- Need to maintain backward compatibility
- Ensure performance doesn't degrade

## Notes
- This is the foundation for all other improvements
- Must be done carefully with extensive testing
- Consider doing in smaller increments if too risky