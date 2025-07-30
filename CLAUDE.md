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

### Running/Testing
```bash
go run cmd/cclogviewer/main.go -input file.jsonl    # Run directly with Go
./bin/cclogviewer -input file.jsonl                 # Run built binary
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
- **internal/models/**: Data structures for log entries and tool calls  
- **internal/parser/**: JSONL file parsing with configurable buffer sizes for large files
- **internal/processor/**: Transforms raw log entries into hierarchical structures, handles tool call matching and sidechain conversation grouping
- **internal/renderer/**: HTML generation with modular templates
  - **templates/**: Base HTML structure and template definitions
  - **templates/styles/**: Modular CSS files (main.css, themes.css, components.css)
  - **templates/scripts/**: JavaScript functionality (main.js)
  - **templates/partials/**: Reusable HTML components (entry.html, tool-call.html)
- **internal/browser/**: Cross-platform browser opening functionality

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