# Task 04: Eliminate Code Duplication

## Priority: 4th (Medium-High)

## Overview
The codebase contains several instances of duplicated code, including entire functions defined in multiple places, repeated logic patterns, and unnecessary wrapper methods. This creates maintenance burden and risk of inconsistencies.

## Issues to Address
1. Duplicate function definitions (setEntryDepth, extractTaskPrompt)
2. Duplicate string extraction logic (GetStringValue vs extractString)
3. Unnecessary wrapper methods in SidechainProcessor
4. Repeated validation patterns across formatters
5. Similar HTML escaping patterns throughout

## Steps to Complete

### Step 1: Remove Duplicate Functions

1. **Remove duplicate setEntryDepth**:
   - Keep version in `internal/processor/hierarchy.go` (lines 25-50)
   - Delete version in `internal/processor/entries.go` (lines 480-505)
   - Update any references to use the HierarchyBuilder method

2. **Remove duplicate extractTaskPrompt**:
   - Keep version in `internal/processor/sidechain.go` (lines 127-143)
   - Delete version in `internal/processor/entries.go` (lines 413-429)
   - Move to a shared utility if needed by multiple packages

3. **Consolidate normalizeText functions**:
   - Check if there are multiple implementations
   - Create single version in utils package
   - Update all callers

### Step 2: Create Shared Utilities Package

Create `internal/utils/extraction.go`:
```go
package utils

// ExtractString safely extracts a string value from a map
func ExtractString(data map[string]interface{}, key string) string {
    if val, ok := data[key]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}

// ExtractBool safely extracts a boolean value from a map
func ExtractBool(data map[string]interface{}, key string) bool {
    if val, ok := data[key]; ok {
        if b, ok := val.(bool); ok {
            return b
        }
    }
    return false
}

// ExtractInt safely extracts an integer value from a map
func ExtractInt(data map[string]interface{}, key string) int {
    if val, ok := data[key]; ok {
        switch v := val.(type) {
        case int:
            return v
        case float64:
            return int(v)
        }
    }
    return 0
}
```

### Step 3: Remove Duplicate String Extraction

1. **Replace all instances of GetStringValue and extractString**:
```go
// Before (in entries.go)
func GetStringValue(m map[string]interface{}, key string) string

// Before (in base.go)
func (b *BaseFormatter) extractString(data map[string]interface{}, key string) string

// After - use shared utility
value := utils.ExtractString(data, "key")
```

2. **Update all callers**:
   - Search for all uses of `GetStringValue`
   - Search for all uses of `extractString`
   - Replace with `utils.ExtractString`

### Step 4: Remove Unnecessary Wrapper Methods

In `internal/processor/sidechain.go`, remove these wrapper methods:
```go
// Delete these methods (lines 270-283):
func (s *SidechainProcessor) collectSidechainEntries(...)
func (s *SidechainProcessor) getFirstUserMessage(...)
func (s *SidechainProcessor) getLastAssistantMessage(...)
```

Update callers to use the standalone functions directly:
```go
// Before
entries := s.collectSidechainEntries(root, entryMap)

// After
entries := collectSidechainEntries(root, entryMap)
```

### Step 5: Create Validation Utilities

Create `internal/utils/validation.go`:
```go
package utils

import "fmt"

// ValidateRequiredFields checks that all required fields are present in the data map
func ValidateRequiredFields(data map[string]interface{}, fields ...string) error {
    var missing []string
    
    for _, field := range fields {
        if ExtractString(data, field) == "" {
            missing = append(missing, field)
        }
    }
    
    if len(missing) > 0 {
        return fmt.Errorf("missing required fields: %v", missing)
    }
    
    return nil
}

// ValidateRequiredField checks that a single required field is present
func ValidateRequiredField(data map[string]interface{}, field string) error {
    if ExtractString(data, field) == "" {
        return fmt.Errorf("missing required field: %s", field)
    }
    return nil
}
```

### Step 6: Refactor Formatter Validation

Update all formatters to use the shared validation:
```go
// Before (repeated in each formatter)
if f.extractString(data, "command") == "" {
    return fmt.Errorf("missing required field: command")
}

// After
if err := utils.ValidateRequiredField(data, "command"); err != nil {
    return err
}

// For multiple fields
if err := utils.ValidateRequiredFields(data, "command", "description", "timeout"); err != nil {
    return err
}
```

### Step 7: Centralize HTML Utilities

Create `internal/utils/html.go`:
```go
package utils

import (
    "html"
    "html/template"
)

// EscapeHTML escapes special HTML characters
func EscapeHTML(s string) string {
    return html.EscapeString(s)
}

// SafeHTML creates a template.HTML from an escaped string
func SafeHTML(s string) template.HTML {
    return template.HTML(EscapeHTML(s))
}

// RawHTML creates a template.HTML without escaping (use with caution)
func RawHTML(s string) template.HTML {
    return template.HTML(s)
}
```

### Step 8: Consolidate Common Patterns

1. **Create a base formatter with common functionality**:
```go
// internal/processor/tools/formatters/enhanced_base.go
type EnhancedBaseFormatter struct {
    BaseFormatter
}

func (f *EnhancedBaseFormatter) ValidateCommonFields(data map[string]interface{}, fields ...string) error {
    return utils.ValidateRequiredFields(data, fields...)
}

func (f *EnhancedBaseFormatter) FormatField(data map[string]interface{}, field string) template.HTML {
    value := utils.ExtractString(data, field)
    if value == "" {
        return template.HTML(`<span class="empty">Not provided</span>`)
    }
    return utils.SafeHTML(value)
}
```

2. **Update formatters to use enhanced base**:
```go
type BashFormatter struct {
    EnhancedBaseFormatter
}

func (f *BashFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
    if err := f.ValidateCommonFields(data, "command"); err != nil {
        return "", err
    }
    
    // Use common formatting
    command := f.FormatField(data, "command")
    // ...
}
```

### Step 9: Remove Duplicate Processing Patterns

Identify and consolidate repeated patterns like:
1. JSON unmarshaling with error checking
2. Message type detection and routing
3. Token calculation logic
4. Tree traversal patterns

Create utility functions for common operations:
```go
// internal/utils/json.go
func UnmarshalToMap(data json.RawMessage) (map[string]interface{}, error) {
    var result map[string]interface{}
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("unmarshaling JSON: %w", err)
    }
    return result, nil
}
```

## Testing Requirements

1. **Before making changes**:
   - Document current behavior
   - Write tests if they don't exist
   - Ensure tests pass

2. **After each refactoring step**:
   - Run all tests
   - Verify no behavior changes
   - Check for performance impact

## Success Criteria
- [ ] No duplicate function definitions
- [ ] Shared utilities package created and used
- [ ] All string extraction uses common utilities
- [ ] Validation patterns consolidated
- [ ] HTML utilities centralized
- [ ] Wrapper methods removed
- [ ] Common patterns extracted to utilities
- [ ] All tests passing
- [ ] No change in functionality

## Migration Strategy

1. Create utilities package first
2. Migrate one component at a time
3. Run tests after each migration
4. Update imports incrementally
5. Remove old code only after new code is tested

## Potential Risks
- Breaking existing functionality
- Import cycles when creating utilities
- Performance regression from indirection

## Notes
- Keep utilities focused and well-documented
- Don't over-abstract - some duplication is OK
- Ensure utilities are truly reusable
- Consider backward compatibility during migration