# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application that converts Claude Code JSONL log files into interactive HTML for easy reading. It's a command-line tool designed to help developers visualize and navigate their Claude Code session logs.

## Common Development Commands

### Building
```bash
make build                    # Build the binary
make build-release           # Build with version info
make build-all              # Build for all platforms (Linux, Darwin, Windows)
```

### Running
```bash
go run cmd/cclogviewer/main.go -input file.jsonl    # Run directly with Go
./bin/cclogviewer -input file.jsonl                 # Run built binary
```

### Testing
```bash
make test                    # Run all unit tests
make test-coverage          # Run tests with coverage report (HTML)
make test-coverage-report   # Show coverage percentage in terminal
make test-integration       # Run integration tests
make test-all              # Run all tests (unit + integration)
make benchmark             # Run benchmarks
```

### Code Quality
```bash
make fmt                     # Format Go code
make lint                    # Run linter (requires golangci-lint)
```

### Installation
```bash
make install                 # Install to /usr/local/bin
make install PREFIX=/opt     # Install to custom prefix
make uninstall              # Remove installed binary
```

### Other Commands
```bash
make clean                   # Clean build artifacts
make deps                    # Download and tidy dependencies
make release                # Create release archives for all platforms
```

## Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

- **cmd/cclogviewer/**: Entry point that handles CLI flags and orchestrates the conversion process
- **internal/browser/**: Cross-platform browser opening functionality
- **internal/constants/**: Centralized constants package containing all magic values organized by category (buffers, thresholds, types, etc.)
- **internal/debug/**: Debug utilities for development
- **internal/models/**: Data structures for log entries and tool calls  
- **internal/parser/**: JSONL file parsing with configurable buffer sizes (removed 10MB line limit)
- **internal/processor/**: Transforms raw log entries into hierarchical structures
  - Modular message handlers for different entry types
  - Tool call matching and result correlation
  - Sidechain (Task tool) conversation grouping
  - **tools/**: Tool processing system with formatter interface
    - **formatters/**: Individual tool formatters (bash, edit, multiedit, read, todowrite, write)
    - **diff/**: Diff computation and formatting utilities
- **internal/renderer/**: HTML generation with modular templates
  - **ansi/**: ANSI color code parsing and conversion
  - **builders/**: HTML building utilities
  - **templates/**: Base HTML structure and template definitions (embedded at compile time)
  - **templates/styles/**: Modular CSS files (main.css, themes.css, components.css)
  - **templates/scripts/**: JavaScript functionality (main.js)
  - **templates/partials/**: Reusable HTML components (entry.html, tool-call.html)
- **internal/testutil/**: Shared test utilities and helpers for consistent testing
- **internal/utils/**: Shared utilities following DRY principles
  - **extraction.go**: JSON extraction utilities (ExtractString, ExtractBool, etc.)
  - **validation.go**: Field validation utilities
  - **html.go**: HTML escaping and formatting utilities
  - **json.go**: JSON manipulation utilities

The processing pipeline:
1. Parse JSONL file into LogEntry structs
2. Process entries to build hierarchical structure and match tool calls with results
3. Group sidechain (Task tool) conversations with their parent tool calls
4. Render processed entries as interactive HTML with expandable sections

Key architectural decisions:
- Uses Go's html/template for safe HTML generation
- Templates are embedded at compile time using Go 1.16+ embed directive
- Modular template structure for maintainability and future enhancements
- Processes entire file in memory for simplicity (suitable for typical log sizes)
- Chronological display with visual hierarchy for nested conversations
- All magic values centralized in constants package for better maintainability
- Shared utilities eliminate code duplication across packages

## Test Requirements (TDD Approach)

This project follows Test-Driven Development (TDD) principles. **Write tests first before implementing features**.

### Test Structure
- **Unit Tests**: Located alongside source files with `_test.go` suffix
- **Integration Tests**: Located in the root directory (`integration_test.go`)
- **Test Utilities**: Available in `internal/testutil/helpers.go`
- **Test Fixtures**: Located in `testdata/fixtures/` with various test cases

### Testing Guidelines
1. **Write Tests First**: Before implementing any new feature or fixing a bug, write a failing test that describes the expected behavior
2. **Use Table-Driven Tests**: Prefer table-driven tests for comprehensive coverage of different scenarios
3. **Use testutil Helpers**: Leverage the helper functions in `internal/testutil` for creating test data
4. **Golden Files**: Use golden files for comparing complex outputs (HTML, formatted text)
5. **Mock External Dependencies**: Use interfaces and dependency injection for testability

### Test Utilities Available
- `GenerateTestUUID()`: Generate UUIDs for testing
- `CreateTestLogEntry()`: Create LogEntry instances
- `CreateTestProcessedEntry()`: Create ProcessedEntry instances
- `CreateToolCallEntry()`: Create tool call entries
- `CreateToolResultEntry()`: Create tool result entries
- `CreateTestToolCallWithResult()`: Create matched tool call/result pairs
- `LoadTestFile()`: Load test fixtures
- `AssertGoldenFile()`: Compare outputs with golden files

### Running Tests
```bash
# Run all tests with verbose output
make test

# Update golden files
UPDATE_GOLDEN=1 make test

# Run benchmarks
make benchmark
```