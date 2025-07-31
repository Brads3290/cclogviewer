# Task 02: Reduce Code Complexity

## Priority: 2nd (High)

## Overview
Many functions in the codebase exceed reasonable complexity limits, with some functions over 100 lines, deep nesting up to 6 levels, and cyclomatic complexity exceeding 15. This makes the code hard to understand, maintain, and test.

## Issues to Address
1. `formatBashResult()` - 97 lines with complex HTML generation
2. `processEntry()` - 80 lines handling multiple message types
3. `extractTaskResult()` - 6 levels of nesting
4. Functions with cyclomatic complexity > 10
5. Long if-else chains and complex switch statements

## Steps to Complete

### Step 1: Refactor formatBashResult (97 lines → <30 lines)
Current location: `internal/renderer/html.go:118-214`

1. Extract output processing logic:
```go
func (r *BashResultFormatter) processOutput(output string) (lines []string, isLong bool) {
    lines = strings.Split(output, "\n")
    isLong = len(lines) > BashOutputCollapseThreshold
    return
}
```

2. Extract HTML generation for collapsible section:
```go
func (r *BashResultFormatter) renderCollapsibleOutput(lines []string, uuid string) string {
    var preview []string
    if len(lines) > 10 {
        preview = append(lines[:5], "...", lines[len(lines)-5:]...)
    }
    // Generate collapsible HTML
    return html
}
```

3. Extract ANSI conversion logic:
```go
func (r *BashResultFormatter) convertLinesToHTML(lines []string) []string {
    converted := make([]string, len(lines))
    for i, line := range lines {
        converted[i] = ConvertANSIToHTML(line)
    }
    return converted
}
```

### Step 2: Simplify processEntry Function
Current location: `internal/processor/entries.go:117-197`

1. Create a message type handler map:
```go
type MessageHandler func(*ProcessingState, *models.ProcessedEntry, map[string]interface{}) error

var messageHandlers = map[string]MessageHandler{
    "text":        handleTextMessage,
    "tool_use":    handleToolUseMessage,
    "tool_result": handleToolResultMessage,
}
```

2. Extract each message type handler:
```go
func handleTextMessage(state *ProcessingState, entry *models.ProcessedEntry, msg map[string]interface{}) error {
    // Handle text message logic
}

func handleToolUseMessage(state *ProcessingState, entry *models.ProcessedEntry, msg map[string]interface{}) error {
    // Handle tool use logic
}
```

3. Simplify main function:
```go
func processEntry(state *ProcessingState, entry *models.ProcessedEntry) {
    var msg map[string]interface{}
    if err := json.Unmarshal(entry.RawMessage, &msg); err != nil {
        log.Printf("Error unmarshaling message: %v", err)
        return
    }
    
    if handler, ok := messageHandlers[entry.Type]; ok {
        if err := handler(state, entry, msg); err != nil {
            log.Printf("Error processing %s message: %v", entry.Type, err)
        }
    }
}
```

### Step 3: Reduce Nesting in extractTaskResult
Current location: `internal/processor/sidechain.go:146-173`

1. Use early returns to reduce nesting:
```go
func extractTaskResult(toolCall *models.ProcessedEntry, originalEntries []models.LogEntry) (string, string) {
    for _, e := range originalEntries {
        if e.UUID != toolCall.Result.UUID {
            continue
        }
        
        result, err := extractResultFromEntry(e)
        if err != nil {
            continue
        }
        
        return result.summary, result.content
    }
    return "", ""
}
```

2. Extract nested logic into separate function:
```go
func extractResultFromEntry(entry models.LogEntry) (*TaskResult, error) {
    var msg map[string]interface{}
    if err := json.Unmarshal(entry.Message, &msg); err != nil {
        return nil, err
    }
    
    content, ok := getContentArray(msg)
    if !ok {
        return nil, fmt.Errorf("no content array")
    }
    
    return extractFromContent(content)
}
```

### Step 4: Simplify Complex Boolean Expressions
1. Extract complex conditions into well-named functions:
```go
// Before
if lcsIdx < len(lcs) && oldIdx < len(oldLines) && newIdx < len(newLines) &&
    oldLines[oldIdx] == lcs[lcsIdx] && newLines[newIdx] == lcs[lcsIdx]

// After
func isMatchingLCSPosition(lcsIdx, oldIdx, newIdx int, lcs, oldLines, newLines []string) bool {
    if lcsIdx >= len(lcs) || oldIdx >= len(oldLines) || newIdx >= len(newLines) {
        return false
    }
    return oldLines[oldIdx] == lcs[lcsIdx] && newLines[newIdx] == lcs[lcsIdx]
}
```

### Step 5: Reduce Function Parameters
For functions with > 5 parameters, create configuration objects:

```go
// Before
func matchTaskWithSidechain(task *models.ProcessedEntry, sidechain *models.Sidechain, 
    entryMap map[string]*models.ProcessedEntry, taskPrompt string, 
    firstUserMessage string, lastAssistantMessage string) float64

// After
type MatchContext struct {
    Task                *models.ProcessedEntry
    Sidechain          *models.Sidechain
    EntryMap           map[string]*models.ProcessedEntry
    TaskPrompt         string
    FirstUserMessage   string
    LastAssistantMessage string
}

func matchTaskWithSidechain(ctx *MatchContext) float64
```

### Step 6: Extract Magic Numbers as Constants
```go
const (
    BashOutputCollapseThreshold = 20
    ShortUUIDLength            = 8
    MinTextLengthForPrefixMatch = 20
    DefaultToolCallMatchWindow  = 5 * time.Minute
)
```

### Step 7: Simplify Long Functions
Guidelines for each long function:
1. Each function should do ONE thing
2. Extract helper functions for distinct operations
3. Use guard clauses for early returns
4. Replace nested ifs with switch or map lookup
5. Extract complex expressions into well-named variables

## Testing Strategy
1. Write tests for each extracted function
2. Ensure behavior remains identical after refactoring
3. Add benchmarks to verify performance isn't degraded

## Success Criteria
- [ ] No function exceeds 50 lines
- [ ] Maximum nesting depth of 3 levels
- [ ] Cyclomatic complexity < 10 for all functions
- [ ] All magic numbers extracted as constants
- [ ] Functions with >5 parameters refactored
- [ ] Complex boolean expressions simplified

## Code Quality Metrics to Track
- Lines per function
- Cyclomatic complexity
- Nesting depth
- Number of parameters per function

## Notes
- Use refactoring tools where available
- Make small, incremental changes
- Run tests after each change
- Consider readability over clever optimizations