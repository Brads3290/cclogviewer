# Task 05: Improve API Design

## Priority: 5th (Medium)

## Overview
The current API design has several issues including inconsistent function signatures, leaking implementation details, missing abstractions, and public APIs that should be private. This task will improve the API to be more consistent, maintainable, and user-friendly.

## Issues to Address
1. Inconsistent function signatures and parameter naming
2. ProcessedEntry exposing 29+ implementation details
3. Missing abstractions (concrete types instead of interfaces)
4. Functions and fields that should be private are exported
5. Poor error handling design
6. Missing builder patterns for complex objects

## Steps to Complete

### Step 1: Create Clean Public API Types

1. **Define minimal public interfaces**:
```go
// internal/interfaces/public.go
package interfaces

// LogEntry represents a parsed log entry (minimal public view)
type LogEntry interface {
    GetUUID() string
    GetType() string
    GetTimestamp() time.Time
    GetContent() string
}

// ProcessedEntry represents a processed entry ready for rendering
type ProcessedEntry interface {
    GetUUID() string
    GetContent() template.HTML
    GetChildren() []ProcessedEntry
    GetDepth() int
    GetTokenCount() int
}

// Processor handles log entry processing
type Processor interface {
    Process(entries []LogEntry) ([]ProcessedEntry, error)
}

// Renderer generates output from processed entries
type Renderer interface {
    Render(entries []ProcessedEntry, options RenderOptions) error
}
```

2. **Create internal implementations**:
```go
// internal/models/internal_entry.go
type internalLogEntry struct {
    // All 22 fields - not exposed
    uuid      string
    entryType string
    // ... other fields
}

// Implement public interface
func (e *internalLogEntry) GetUUID() string { return e.uuid }
func (e *internalLogEntry) GetType() string { return e.entryType }
```

### Step 2: Implement Builder Pattern

1. **Create builders for complex objects**:
```go
// internal/builders/entry_builder.go
package builders

type ProcessedEntryBuilder struct {
    entry *models.InternalProcessedEntry
}

func NewProcessedEntryBuilder() *ProcessedEntryBuilder {
    return &ProcessedEntryBuilder{
        entry: &models.InternalProcessedEntry{},
    }
}

func (b *ProcessedEntryBuilder) WithUUID(uuid string) *ProcessedEntryBuilder {
    b.entry.UUID = uuid
    return b
}

func (b *ProcessedEntryBuilder) WithContent(content string) *ProcessedEntryBuilder {
    b.entry.Content = content
    return b
}

func (b *ProcessedEntryBuilder) WithDepth(depth int) *ProcessedEntryBuilder {
    b.entry.Depth = depth
    return b
}

func (b *ProcessedEntryBuilder) Build() (interfaces.ProcessedEntry, error) {
    // Validate required fields
    if b.entry.UUID == "" {
        return nil, errors.New("UUID is required")
    }
    return b.entry, nil
}
```

2. **Create configuration builders**:
```go
// ProcessorConfig using builder pattern
type ProcessorConfigBuilder struct {
    config *ProcessorConfig
}

func NewProcessorConfig() *ProcessorConfigBuilder {
    return &ProcessorConfigBuilder{
        config: &ProcessorConfig{
            WindowSize:  5 * time.Minute, // default
            BufferSize:  64 * 1024,      // default
            DebugMode:   false,          // default
        },
    }
}

func (b *ProcessorConfigBuilder) WithDebugMode(debug bool) *ProcessorConfigBuilder {
    b.config.DebugMode = debug
    return b
}

func (b *ProcessorConfigBuilder) WithWindowSize(d time.Duration) *ProcessorConfigBuilder {
    b.config.WindowSize = d
    return b
}

func (b *ProcessorConfigBuilder) Build() *ProcessorConfig {
    return b.config
}
```

### Step 3: Standardize Function Signatures

1. **Create consistent parameter ordering**:
```go
// Standardize: context/config first, main data second, options last

// Before (inconsistent)
func ReadJSONLFile(filename string) ([]LogEntry, error)
func GenerateHTML(entries []ProcessedEntry, outputFile string, debugMode bool) error

// After (consistent)
func ReadJSONLFile(ctx context.Context, filepath string, opts ...ReadOption) ([]LogEntry, error)
func GenerateHTML(ctx context.Context, entries []ProcessedEntry, opts ...RenderOption) error
```

2. **Use functional options pattern**:
```go
type ReadOption func(*readOptions)

type readOptions struct {
    bufferSize int
    maxLineSize int
}

func WithBufferSize(size int) ReadOption {
    return func(o *readOptions) {
        o.bufferSize = size
    }
}

func WithMaxLineSize(size int) ReadOption {
    return func(o *readOptions) {
        o.maxLineSize = size
    }
}

// Usage:
entries, err := ReadJSONLFile(ctx, "file.jsonl", 
    WithBufferSize(128*1024),
    WithMaxLineSize(20*1024*1024),
)
```

### Step 4: Fix Error Handling

1. **Create custom error types**:
```go
// internal/errors/errors.go
package errors

type ErrorType string

const (
    ErrorTypeParsing    ErrorType = "parsing"
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeProcessing ErrorType = "processing"
    ErrorTypeIO         ErrorType = "io"
)

type ProcessingError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *ProcessingError) Error() string {
    return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

func (e *ProcessingError) Unwrap() error {
    return e.Cause
}

// Helper constructors
func NewParsingError(message string, cause error) *ProcessingError {
    return &ProcessingError{
        Type:    ErrorTypeParsing,
        Message: message,
        Cause:   cause,
    }
}
```

2. **Improve error context**:
```go
// Before
t, _ := time.Parse(time.RFC3339, ts)

// After
func parseTimestamp(ts string) (time.Time, error) {
    t, err := time.Parse(time.RFC3339, ts)
    if err != nil {
        return time.Time{}, NewParsingError(
            fmt.Sprintf("invalid timestamp format: %q", ts),
            err,
        )
    }
    return t, nil
}
```

### Step 5: Hide Implementation Details

1. **Make internal functions unexported**:
```go
// Before (exported but only used internally)
func GetStringValue(m map[string]interface{}, key string) string
func GetBoolValue(m map[string]interface{}, key string) bool

// After (unexported)
func getStringValue(m map[string]interface{}, key string) string
func getBoolValue(m map[string]interface{}, key string) bool
```

2. **Use internal packages**:
```
internal/
├── api/              # Public API implementations
├── core/             # Internal business logic
│   └── internal/     # Really internal stuff
├── utils/            # Shared utilities
│   └── internal/     # Internal utilities
```

### Step 6: Create Facade for Complex Subsystems

```go
// internal/api/facade.go
package api

// CCLogViewer provides a simplified API for the entire system
type CCLogViewer struct {
    parser    Parser
    processor Processor
    renderer  Renderer
}

// New creates a new CCLogViewer instance with default configuration
func New() *CCLogViewer {
    return &CCLogViewer{
        parser:    parser.New(),
        processor: processor.New(),
        renderer:  renderer.New(),
    }
}

// ConvertFile is the main entry point for file conversion
func (c *CCLogViewer) ConvertFile(inputPath, outputPath string, opts ...Option) error {
    // Orchestrate the entire pipeline
    entries, err := c.parser.ParseFile(inputPath)
    if err != nil {
        return fmt.Errorf("parsing file: %w", err)
    }
    
    processed, err := c.processor.Process(entries)
    if err != nil {
        return fmt.Errorf("processing entries: %w", err)
    }
    
    if err := c.renderer.RenderToFile(processed, outputPath); err != nil {
        return fmt.Errorf("rendering output: %w", err)
    }
    
    return nil
}
```

### Step 7: Define Clear Interfaces

1. **Create interface segregation**:
```go
// Instead of one big ToolFormatter interface
type InputFormatter interface {
    FormatInput(data map[string]interface{}) (template.HTML, error)
}

type InputValidator interface {
    ValidateInput(data map[string]interface{}) error
}

type DescriptionProvider interface {
    GetDescription(data map[string]interface{}) string
}

// Formatters can implement multiple interfaces
type BashFormatter struct {
    // ...
}

// Implement all three interfaces
func (f *BashFormatter) FormatInput(...) (template.HTML, error) { }
func (f *BashFormatter) ValidateInput(...) error { }
func (f *BashFormatter) GetDescription(...) string { }
```

### Step 8: Improve Type Safety

1. **Replace map[string]interface{} with typed structs**:
```go
// Before
func FormatInput(data map[string]interface{}) (template.HTML, error)

// After
type BashInput struct {
    Command     string
    Description string
    Timeout     *time.Duration
    CWD         string
}

func FormatInput(input *BashInput) (template.HTML, error)
```

2. **Use type aliases for clarity**:
```go
type UUID string
type SessionID string
type EntryType string

const (
    EntryTypeMessage    EntryType = "message"
    EntryTypeToolCall   EntryType = "tool_call"
    EntryTypeToolResult EntryType = "tool_result"
)
```

## Testing Strategy

1. Create comprehensive tests for all public APIs
2. Test builders with various configurations
3. Test error types and error handling
4. Verify interfaces are properly implemented

## Success Criteria
- [ ] Clean separation between public API and implementation
- [ ] All complex objects use builder pattern
- [ ] Consistent function signatures throughout
- [ ] Custom error types with proper context
- [ ] Implementation details hidden behind interfaces
- [ ] Functional options pattern for configuration
- [ ] Type safety improved with proper structs
- [ ] Clear, minimal public API surface

## Migration Plan

1. Create new API alongside existing code
2. Implement adapters to existing implementation
3. Migrate consumers one at a time
4. Deprecate old APIs
5. Remove old code after migration complete

## Notes
- Keep interfaces small and focused
- Document behavior contracts clearly
- Consider backward compatibility
- Use semantic versioning for API changes