# Claude Code Log Viewer

A Go program that converts Claude Code JSONL log files into beautiful, interactive HTML for easy reading.

## Features

- 📖 Converts JSONL logs to clean HTML interface
- 🎯 Hierarchical conversation display with proper nesting
- 📎 Expandable Task tool calls showing nested execution
- 🔧 Expandable tool call details with formatted inputs
- 🎨 Color-coded messages (user/assistant)
- ⏱️ Timestamps for each message
- 🔍 Handles large log files efficiently
- 🚀 Quick view mode - automatically opens in browser when output is omitted

## Usage

```bash
# Quick view - generates temp file and opens automatically
./cclogviewer -input /path/to/logfile.jsonl

# Specify output file
./cclogviewer -input /path/to/logfile.jsonl -output conversation.html

# Specify output and open in browser
./cclogviewer -input /path/to/logfile.jsonl -output conversation.html -open
```

Or run directly with Go:

```bash
go run main.go -input /path/to/logfile.jsonl
```

## Arguments

- `-input`: Path to the Claude Code JSONL log file (required)
- `-output`: Path for the output HTML file (optional - if omitted, creates temp file and auto-opens)
- `-open`: Force open the generated HTML file in browser (automatic when output is omitted)

## Examples

```bash
# Quick view - generates temp file and opens automatically
./cclogviewer -input ~/.claude/projects/myproject/session.jsonl

# Save to specific file without opening
./cclogviewer -input session.jsonl -output myproject.html

# Save to specific file and open
./cclogviewer -input session.jsonl -output myproject.html -open
```

When output is omitted, files are saved to `/tmp/` with names like `cclog-session-20250727-143052.html`

## Features in the HTML

- Click on tool calls to expand/collapse their details
- Task tool calls show nested conversations within
- Error messages are highlighted in red
- Code blocks are properly formatted
- Sidechain (Task) conversations are visually distinguished