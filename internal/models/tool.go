package models

import "html/template"

// ToolCall represents a tool invocation with its result.
type ToolCall struct {
	ID                  string
	Name                string
	Description         string
	Input               template.HTML
	RawInput            interface{}       // Raw input data before formatting
	CompactView         template.HTML     // Optional compact view for specific tools
	Result              *ProcessedEntry   // Tool result entry
	TaskEntries         []*ProcessedEntry // For Task tool - sidechain entries
	IsInterrupted       bool              // Whether the tool was interrupted by the user
	HasMissingResult    bool              // Whether the tool result is missing
	HasMissingSidechain bool              // Whether Task tool sidechain conversation is missing
	CWD                 string            // Current working directory when the tool was called
}
