package models

import (
	"encoding/json"
	"html/template"
)

// LogEntry represents a single line from the JSONL file
type LogEntry struct {
	ParentUUID   *string         `json:"parentUuid"`
	IsSidechain  bool            `json:"isSidechain"`
	UserType     string          `json:"userType"`
	CWD          string          `json:"cwd"`
	SessionID    string          `json:"sessionId"`
	Version      string          `json:"version"`
	GitBranch    string          `json:"gitBranch"`
	Type         string          `json:"type"`
	Message      json.RawMessage `json:"message"`
	RequestID    string          `json:"requestId"`
	UUID         string          `json:"uuid"`
	Timestamp    string          `json:"timestamp"`
	IsMeta       bool            `json:"isMeta"`
	ToolUseResult interface{}    `json:"toolUseResult"`
}

// ProcessedEntry represents a processed log entry for display
type ProcessedEntry struct {
	UUID         string
	ParentUUID   string
	IsSidechain  bool
	Type         string
	Timestamp    string
	RawTimestamp string // Keep the raw timestamp for comparisons
	Role         string
	Content      template.HTML
	ToolCalls    []ToolCall
	IsToolResult bool
	IsError      bool
	Children     []*ProcessedEntry
	Depth        int
	ToolResultID string // For matching tool results to tool calls
}