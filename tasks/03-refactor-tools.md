# Task 03: Refactor tools.go

**File:** `internal/processor/tools.go`  
**Current Size:** 497 lines  
**Priority:** MEDIUM  
**Estimated Effort:** 4-5 hours  

## Problem Summary

The `tools.go` file contains all tool-specific formatting logic in a single file with:
- Long functions (several over 50 lines)
- Repetitive formatting patterns across different tools
- Complex diff algorithms mixed with presentation logic
- Embedded HTML generation throughout
- No clear abstraction for adding new tool types

## Current State Analysis

### Key Issues

1. **Monolithic Tool Formatting:**
   - Single massive `switch` statement in `formatToolInput()` (lines 29-124)
   - Each tool formatter is a large function with mixed concerns
   - No interface or abstraction for tool formatters

2. **Repetitive Patterns:**
   ```go
   // This pattern repeats for each tool
   func formatEditTool(data map[string]interface{}) string {
       // Extract parameters
       // Validate data
       // Generate HTML
       // Return formatted string
   }
   ```

3. **Complex Diff Logic:**
   - `computeLineDiff()` is 80+ lines of complex algorithm
   - `formatUnifiedDiff()` mixes diff computation with HTML generation
   - Diff visualization logic tightly coupled to Edit tool

4. **HTML Generation Mixed with Logic:**
   - HTML strings embedded throughout
   - No template usage
   - Difficult to change styling or structure

## Proposed Solution

### Architecture Redesign

```
processor/tools/
├── formatter.go        # Main formatter interface and registry
├── formatters/
│   ├── base.go        # Base formatter with common functionality
│   ├── edit.go        # Edit tool formatter
│   ├── write.go       # Write tool formatter
│   ├── read.go        # Read tool formatter
│   ├── bash.go        # Bash tool formatter
│   └── ...            # Other tool formatters
├── diff/
│   ├── compute.go     # Diff computation algorithms
│   ├── format.go      # Diff formatting utilities
│   └── models.go      # Diff-related data structures
└── templates/
    └── tool.html      # Tool display templates
```

### Key Abstractions

1. **ToolFormatter Interface:**
   ```go
   type ToolFormatter interface {
       Name() string
       FormatInput(data map[string]interface{}) (string, error)
       FormatOutput(output interface{}) (string, error)
       ValidateInput(data map[string]interface{}) error
   }
   ```

2. **FormatterRegistry:**
   ```go
   type FormatterRegistry struct {
       formatters map[string]ToolFormatter
   }
   
   func (r *FormatterRegistry) Register(formatter ToolFormatter) {
       r.formatters[formatter.Name()] = formatter
   }
   
   func (r *FormatterRegistry) Format(toolName string, data map[string]interface{}) (string, error) {
       formatter, exists := r.formatters[toolName]
       if !exists {
           return r.formatGeneric(toolName, data)
       }
       return formatter.FormatInput(data)
   }
   ```

3. **Base Formatter:**
   ```go
   type BaseFormatter struct {
       toolName string
       template *template.Template
   }
   
   func (b *BaseFormatter) extractString(data map[string]interface{}, key string) string {
       // Common parameter extraction logic
   }
   
   func (b *BaseFormatter) renderTemplate(name string, data interface{}) (string, error) {
       // Common template rendering
   }
   ```

## Implementation Steps

### Step 1: Create Formatter Interface and Registry (1 hour)

Create `internal/processor/tools/formatter.go`:

```go
package tools

import (
    "fmt"
    "html/template"
)

// ToolFormatter defines the interface for tool-specific formatters
type ToolFormatter interface {
    Name() string
    FormatInput(data map[string]interface{}) (string, error)
    FormatOutput(output interface{}) (string, error)
    ValidateInput(data map[string]interface{}) error
}

// FormatterRegistry manages tool formatters
type FormatterRegistry struct {
    formatters map[string]ToolFormatter
    mu         sync.RWMutex
}

func NewFormatterRegistry() *FormatterRegistry {
    return &FormatterRegistry{
        formatters: make(map[string]ToolFormatter),
    }
}

func (r *FormatterRegistry) Register(formatter ToolFormatter) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.formatters[formatter.Name()] = formatter
}

func (r *FormatterRegistry) Format(toolName string, data map[string]interface{}) (string, error) {
    r.mu.RLock()
    formatter, exists := r.formatters[toolName]
    r.mu.RUnlock()
    
    if !exists {
        return r.formatGeneric(toolName, data)
    }
    
    if err := formatter.ValidateInput(data); err != nil {
        return "", fmt.Errorf("invalid input for %s: %w", toolName, err)
    }
    
    return formatter.FormatInput(data)
}
```

### Step 2: Create Base Formatter (45 min)

Create `internal/processor/tools/formatters/base.go`:

```go
package formatters

type BaseFormatter struct {
    toolName string
}

// Common utility methods
func (b *BaseFormatter) Name() string {
    return b.toolName
}

func (b *BaseFormatter) extractString(data map[string]interface{}, key string) string {
    if val, ok := data[key]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}

func (b *BaseFormatter) extractBool(data map[string]interface{}, key string) bool {
    if val, ok := data[key]; ok {
        if b, ok := val.(bool); ok {
            return b
        }
    }
    return false
}

func (b *BaseFormatter) escapeHTML(s string) string {
    return html.EscapeString(s)
}

func (b *BaseFormatter) formatPath(path string) string {
    return fmt.Sprintf(`<span class="file-path">%s</span>`, b.escapeHTML(path))
}
```

### Step 3: Extract Diff Utilities (1 hour)

Create `internal/processor/tools/diff/compute.go`:

```go
package diff

type DiffLine struct {
    Type     LineType // Add, Remove, Context
    Number   int
    Content  string
}

type LineType int

const (
    LineContext LineType = iota
    LineAdd
    LineRemove
)

// ComputeLineDiff computes the difference between two strings
func ComputeLineDiff(oldStr, newStr string) []DiffLine {
    // Extract the diff computation algorithm from tools.go
    // This is pure logic, no HTML generation
}

// ComputeUnifiedDiff generates a unified diff format
func ComputeUnifiedDiff(oldStr, newStr string, context int) string {
    // Extract unified diff logic
}
```

Create `internal/processor/tools/diff/format.go`:

```go
package diff

// FormatDiffHTML formats diff lines as HTML
func FormatDiffHTML(lines []DiffLine) string {
    var result strings.Builder
    
    for _, line := range lines {
        switch line.Type {
        case LineAdd:
            result.WriteString(formatAddLine(line))
        case LineRemove:
            result.WriteString(formatRemoveLine(line))
        case LineContext:
            result.WriteString(formatContextLine(line))
        }
    }
    
    return result.String()
}
```

### Step 4: Implement Tool-Specific Formatters (1.5 hours)

Create individual formatter files for each tool:

**`internal/processor/tools/formatters/edit.go`:**
```go
package formatters

type EditFormatter struct {
    BaseFormatter
}

func NewEditFormatter() *EditFormatter {
    return &EditFormatter{
        BaseFormatter: BaseFormatter{toolName: "Edit"},
    }
}

func (f *EditFormatter) FormatInput(data map[string]interface{}) (string, error) {
    filePath := f.extractString(data, "file_path")
    oldString := f.extractString(data, "old_string")
    newString := f.extractString(data, "new_string")
    
    // Use diff package for diff computation
    diffLines := diff.ComputeLineDiff(oldString, newString)
    
    // Use template for HTML generation
    return f.renderEditTemplate(EditData{
        FilePath:  filePath,
        DiffLines: diffLines,
    })
}

func (f *EditFormatter) ValidateInput(data map[string]interface{}) error {
    required := []string{"file_path", "old_string", "new_string"}
    for _, field := range required {
        if f.extractString(data, field) == "" {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    return nil
}
```

**Similar pattern for other tools:**
- `write.go` - Write tool formatter
- `read.go` - Read tool formatter  
- `bash.go` - Bash tool formatter
- etc.

### Step 5: Update Main Tools Processing (45 min)

Refactor the main tools.go file:

```go
package processor

import (
    "github.com/user/cclogviewer/internal/processor/tools"
    "github.com/user/cclogviewer/internal/processor/tools/formatters"
)

var registry *tools.FormatterRegistry

func init() {
    registry = tools.NewFormatterRegistry()
    
    // Register all formatters
    registry.Register(formatters.NewEditFormatter())
    registry.Register(formatters.NewWriteFormatter())
    registry.Register(formatters.NewReadFormatter())
    registry.Register(formatters.NewBashFormatter())
    // ... register other formatters
}

// FormatToolCall formats a tool call for display
func FormatToolCall(toolCall *models.ToolCall) (*FormattedTool, error) {
    formatted := &FormattedTool{
        ID:   toolCall.ID,
        Name: toolCall.Name,
    }
    
    // Use registry for formatting
    input, err := registry.Format(toolCall.Name, toolCall.Input)
    if err != nil {
        return nil, fmt.Errorf("formatting input: %w", err)
    }
    formatted.Input = input
    
    return formatted, nil
}
```

### Step 6: Add Tests (45 min)

Create comprehensive tests:
- `formatter_test.go` - Test formatter registry
- `formatters/edit_test.go` - Test edit formatter
- `diff/compute_test.go` - Test diff algorithms
- Integration tests for the refactored system

## Benefits

1. **Extensibility:** Easy to add new tool formatters
2. **Maintainability:** Each formatter is self-contained
3. **Testability:** Small, focused components
4. **Reusability:** Common logic in base formatter
5. **Separation of Concerns:** Diff logic separate from formatting

## Risks and Mitigation

1. **Risk:** Breaking existing tool formatting
   - **Mitigation:** Comprehensive test coverage before refactoring
   - **Mitigation:** A/B test outputs to ensure identical results

2. **Risk:** Performance overhead from abstraction
   - **Mitigation:** Benchmark formatter performance
   - **Mitigation:** Use sync.Pool for template execution if needed

3. **Risk:** Increased complexity for simple tools
   - **Mitigation:** Provide sensible defaults in base formatter
   - **Mitigation:** Keep simple tools simple

## Success Criteria

- [ ] All tool formatters extracted to separate files
- [ ] No formatter exceeds 100 lines
- [ ] Diff logic completely separated from formatting
- [ ] Easy to add new tool types (< 50 lines of code)
- [ ] 90%+ test coverage for formatters
- [ ] No visual changes to output

## Future Enhancements

This refactoring enables:
- Plugin system for custom tool formatters
- Configurable output formats (HTML, Markdown, Plain text)
- Tool-specific syntax highlighting
- Custom diff algorithms per tool type
- Internationalization of tool displays