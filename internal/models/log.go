package models

import (
	"encoding/json"
)

// LogEntry represents a single line from the JSONL file
type LogEntry struct {
	ParentUUID    *string         `json:"parentUuid"`
	IsSidechain   bool            `json:"isSidechain"`
	UserType      string          `json:"userType"`
	CWD           string          `json:"cwd"`
	SessionID     string          `json:"sessionId"`
	Version       string          `json:"version"`
	GitBranch     string          `json:"gitBranch"`
	Type          string          `json:"type"`
	Message       json.RawMessage `json:"message"`
	RequestID     string          `json:"requestId"`
	UUID          string          `json:"uuid"`
	Timestamp     string          `json:"timestamp"`
	IsMeta        bool            `json:"isMeta"`
	ToolUseResult interface{}     `json:"toolUseResult"`
}

// ProcessedEntry represents a processed log entry for display
type ProcessedEntry struct {
	UUID                string
	ParentUUID          string
	IsSidechain         bool
	Type                string
	Timestamp           string
	RawTimestamp        string // Keep the raw timestamp for comparisons
	Role                string
	Content             string // Raw content, HTML escaping happens in templates
	ToolCalls           []ToolCall
	IsToolResult        bool
	IsError             bool
	Children            []*ProcessedEntry
	Depth               int
	ToolResultID        string // For matching tool results to tool calls
	TokenCount          int    // Tokens in this message (output tokens for assistant, estimated for user)
	TotalTokens         int    // Running total of all tokens up to this message
	InputTokens         int    // Input tokens from usage
	OutputTokens        int    // Output tokens from usage
	CacheReadTokens     int    // Cache read tokens from usage
	CacheCreationTokens int    // Cache creation tokens from usage
	IsCaveatMessage     bool   // True if this is a special caveat message from local commands
	IsCommandMessage    bool   // True if this is a command message with XML syntax
	CommandName         string // The command name (e.g., "/add-dir")
	CommandArgs         string // The command arguments
	CommandOutput       string // The stdout output from the command
}
