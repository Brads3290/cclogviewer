# Task 07: Add Comprehensive Testing

## Priority: 7th (High)

## Overview
The project currently has critically low test coverage at 6.2%, with most core components having 0% coverage. This task will establish a comprehensive testing strategy and implement tests for all critical paths.

## Current State
- Overall coverage: 6.2%
- Only 2 test files exist
- 14 packages have zero test coverage
- No integration tests
- No test fixtures or test data

## Target State
- Minimum 50% coverage for Phase 1
- 80% coverage for critical paths
- Full integration test suite
- Comprehensive test fixtures
- Performance benchmarks

## Steps to Complete

### Step 1: Set Up Testing Infrastructure

1. **Create test data directory structure**:
```bash
testdata/
├── fixtures/
│   ├── valid/
│   │   ├── simple.jsonl
│   │   ├── with_tools.jsonl
│   │   ├── with_sidechain.jsonl
│   │   └── complex_nested.jsonl
│   ├── invalid/
│   │   ├── malformed.jsonl
│   │   ├── empty.jsonl
│   │   └── corrupted.jsonl
│   └── edge_cases/
│       ├── huge_file.jsonl
│       ├── unicode.jsonl
│       └── special_chars.jsonl
├── golden/
│   ├── simple.html
│   └── with_tools.html
└── benchmarks/
    └── large_conversation.jsonl
```

2. **Create test helpers package**:
```go
// internal/testutil/helpers.go
package testutil

import (
    "encoding/json"
    "testing"
    "time"
)

// CreateTestLogEntry creates a LogEntry for testing
func CreateTestLogEntry(t *testing.T, entryType, content string) *models.LogEntry {
    t.Helper()
    
    message, err := json.Marshal(map[string]interface{}{
        "content": content,
    })
    require.NoError(t, err)
    
    return &models.LogEntry{
        UUID:      GenerateTestUUID(),
        Type:      entryType,
        Timestamp: time.Now().Format(time.RFC3339),
        Message:   message,
    }
}

// LoadTestFile loads a test fixture file
func LoadTestFile(t *testing.T, path string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", path))
    require.NoError(t, err)
    return data
}

// AssertGoldenFile compares output with golden file
func AssertGoldenFile(t *testing.T, actual []byte, goldenPath string) {
    t.Helper()
    // Implementation
}
```

### Step 2: Test Core Parser Package

**Create**: `internal/parser/jsonl_test.go`

```go
package parser

import (
    "testing"
    "path/filepath"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestReadJSONLFile(t *testing.T) {
    tests := []struct {
        name      string
        file      string
        wantCount int
        wantErr   bool
    }{
        {
            name:      "valid simple file",
            file:      "fixtures/valid/simple.jsonl",
            wantCount: 10,
            wantErr:   false,
        },
        {
            name:      "empty file",
            file:      "fixtures/invalid/empty.jsonl",
            wantCount: 0,
            wantErr:   false,
        },
        {
            name:      "malformed JSON",
            file:      "fixtures/invalid/malformed.jsonl",
            wantCount: 5, // Should skip bad lines
            wantErr:   false,
        },
        {
            name:    "non-existent file",
            file:    "does_not_exist.jsonl",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            path := filepath.Join("testdata", tt.file)
            entries, err := ReadJSONLFile(path)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Len(t, entries, tt.wantCount)
        })
    }
}

func TestReadJSONLFile_LargeFile(t *testing.T) {
    // Test with file larger than buffer size
    entries, err := ReadJSONLFile("testdata/benchmarks/large_conversation.jsonl")
    require.NoError(t, err)
    assert.Greater(t, len(entries), 1000)
}

func TestReadJSONLFile_SpecialCharacters(t *testing.T) {
    entries, err := ReadJSONLFile("testdata/edge_cases/unicode.jsonl")
    require.NoError(t, err)
    
    // Verify Unicode content is preserved
    for _, entry := range entries {
        assert.Contains(t, string(entry.Message), "🚀")
    }
}

func BenchmarkReadJSONLFile(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, _ = ReadJSONLFile("testdata/benchmarks/large_conversation.jsonl")
    }
}
```

### Step 3: Test Entry Processing

**Create**: `internal/processor/entries_test.go`

```go
package processor

import (
    "testing"
    "github.com/brads3290/cclogviewer/internal/models"
    "github.com/brads3290/cclogviewer/internal/testutil"
)

func TestProcessEntries(t *testing.T) {
    tests := []struct {
        name            string
        entries         []models.LogEntry
        wantRootCount   int
        wantTotalTokens int
    }{
        {
            name: "simple conversation",
            entries: []models.LogEntry{
                *testutil.CreateTestLogEntry(t, "message", "Hello"),
                *testutil.CreateTestLogEntry(t, "message", "Hi there"),
            },
            wantRootCount: 2,
        },
        {
            name: "with tool calls",
            entries: createToolCallEntries(t),
            wantRootCount: 1, // Tool result should be nested
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ProcessEntries(tt.entries)
            assert.Len(t, result, tt.wantRootCount)
            
            // Verify hierarchy
            for _, entry := range result {
                assert.NotNil(t, entry)
                assert.Greater(t, entry.Depth, 0)
            }
        })
    }
}

func TestProcessEntry_MessageTypes(t *testing.T) {
    // Test each message type processing
    types := []string{"text", "tool_use", "tool_result"}
    
    for _, msgType := range types {
        t.Run(msgType, func(t *testing.T) {
            entry := createTestEntryOfType(t, msgType)
            state := NewProcessingState()
            
            processEntry(state, entry)
            
            // Verify appropriate fields are set
            assert.NotEmpty(t, entry.Content)
            assert.Greater(t, entry.TokenCount, 0)
        })
    }
}

func TestProcessEntry_ErrorHandling(t *testing.T) {
    // Test with malformed JSON
    entry := &models.ProcessedEntry{
        RawMessage: []byte(`{invalid json`),
    }
    
    state := NewProcessingState()
    processEntry(state, entry)
    
    // Should not panic, should handle gracefully
    assert.Empty(t, entry.Content)
}
```

### Step 4: Test Tool Call Matching

**Create**: `internal/processor/matcher_test.go`

```go
package processor

import (
    "testing"
    "time"
)

func TestToolCallMatcher_MatchToolCalls(t *testing.T) {
    matcher := NewToolCallMatcher()
    
    t.Run("matches within time window", func(t *testing.T) {
        state := createStateWithToolCallAndResult(t, 1*time.Minute)
        err := matcher.MatchToolCalls(state)
        require.NoError(t, err)
        
        // Verify match
        toolCall := findToolCall(state, "test-tool")
        assert.NotNil(t, toolCall.Result)
    })
    
    t.Run("no match outside time window", func(t *testing.T) {
        state := createStateWithToolCallAndResult(t, 10*time.Minute)
        err := matcher.MatchToolCalls(state)
        require.NoError(t, err)
        
        // Verify no match
        toolCall := findToolCall(state, "test-tool")
        assert.Nil(t, toolCall.Result)
    })
    
    t.Run("handles interrupted tool calls", func(t *testing.T) {
        state := createInterruptedToolCall(t)
        err := matcher.MatchToolCalls(state)
        require.NoError(t, err)
        
        // Verify proper handling
        toolCall := findToolCall(state, "interrupted-tool")
        assert.True(t, toolCall.WasInterrupted)
    })
}
```

### Step 5: Test HTML Rendering

**Create**: `internal/renderer/html_test.go`

```go
package renderer

import (
    "testing"
    "bytes"
    "strings"
)

func TestGenerateHTML(t *testing.T) {
    entries := createTestProcessedEntries(t)
    
    t.Run("generates valid HTML", func(t *testing.T) {
        var buf bytes.Buffer
        err := GenerateHTMLToWriter(entries, &buf, false)
        require.NoError(t, err)
        
        html := buf.String()
        assert.Contains(t, html, "<!DOCTYPE html>")
        assert.Contains(t, html, "</html>")
        
        // Verify entries are rendered
        for _, entry := range entries {
            assert.Contains(t, html, entry.UUID)
        }
    })
    
    t.Run("escapes HTML in content", func(t *testing.T) {
        entries := []*models.ProcessedEntry{{
            Content: "<script>alert('xss')</script>",
        }}
        
        var buf bytes.Buffer
        err := GenerateHTMLToWriter(entries, &buf, false)
        require.NoError(t, err)
        
        html := buf.String()
        assert.NotContains(t, html, "<script>alert")
        assert.Contains(t, html, "&lt;script&gt;")
    })
}

func TestConvertANSIToHTML(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {
            name:  "basic color",
            input: "\x1b[31mRed Text\x1b[0m",
            want:  `<span style="color: #cc0000">Red Text</span>`,
        },
        {
            name:  "no ANSI codes",
            input: "Plain text",
            want:  "Plain text",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ConvertANSIToHTML(tt.input)
            assert.Equal(t, tt.want, string(got))
        })
    }
}
```

### Step 6: Integration Tests

**Create**: `internal/integration_test.go`

```go
package cclogviewer_test

import (
    "testing"
    "os"
    "path/filepath"
)

func TestEndToEnd_SimpleConversion(t *testing.T) {
    // Read test JSONL
    inputPath := "testdata/fixtures/valid/simple.jsonl"
    outputPath := filepath.Join(t.TempDir(), "output.html")
    
    // Run full pipeline
    err := ConvertFile(inputPath, outputPath, false)
    require.NoError(t, err)
    
    // Verify output exists
    _, err = os.Stat(outputPath)
    require.NoError(t, err)
    
    // Compare with golden file
    testutil.AssertGoldenFile(t, outputPath, "golden/simple.html")
}

func TestEndToEnd_ComplexConversation(t *testing.T) {
    testCases := []string{
        "with_tools.jsonl",
        "with_sidechain.jsonl", 
        "complex_nested.jsonl",
    }
    
    for _, tc := range testCases {
        t.Run(tc, func(t *testing.T) {
            inputPath := filepath.Join("testdata/fixtures/valid", tc)
            outputPath := filepath.Join(t.TempDir(), "output.html")
            
            err := ConvertFile(inputPath, outputPath, false)
            require.NoError(t, err)
            
            // Verify specific features
            content, _ := os.ReadFile(outputPath)
            verifyHTMLContent(t, content, tc)
        })
    }
}
```

### Step 7: Create Test Fixtures

**Create**: `testdata/fixtures/valid/simple.jsonl`
```json
{"uuid":"1","type":"message","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"uuid":"2","type":"message","timestamp":"2024-01-01T10:00:01Z","message":{"role":"assistant","content":"Hi there!"}}
```

**Create**: `testdata/fixtures/valid/with_tools.jsonl`
```json
{"uuid":"1","type":"message","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Run ls command"}}
{"uuid":"2","type":"tool_call","timestamp":"2024-01-01T10:00:01Z","message":{"name":"Bash","input":{"command":"ls"}}}
{"uuid":"3","type":"tool_result","timestamp":"2024-01-01T10:00:02Z","tool_use_id":"2","message":{"output":"file1.txt\nfile2.txt"}}
```

### Step 8: Add Test Coverage Targets

**Update**: `Makefile`
```makefile
.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-coverage-report
test-coverage-report: test-coverage
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

.PHONY: test-integration
test-integration:
	go test -tags=integration -v ./...

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...
```

### Step 9: Add CI/CD Test Configuration

**Create**: `.github/workflows/test.yml`
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: make test
    
    - name: Check coverage
      run: |
        make test-coverage-report
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Coverage: $coverage%"
        if (( $(echo "$coverage < 50" | bc -l) )); then
          echo "Coverage is below 50%"
          exit 1
        fi
```

## Success Criteria
- [ ] Test coverage > 50% overall
- [ ] Core packages have > 80% coverage
- [ ] All critical paths have tests
- [ ] Integration tests pass
- [ ] No untested error paths
- [ ] Benchmarks established
- [ ] CI/CD pipeline includes tests

## Testing Best Practices
1. Use table-driven tests
2. Test both happy path and error cases
3. Use test fixtures for complex data
4. Mock external dependencies
5. Keep tests fast and isolated
6. Use meaningful test names

## Notes
- Start with critical path testing
- Add tests before fixing bugs
- Use coverage as a guide, not a goal
- Focus on behavior, not implementation
- Keep test data realistic