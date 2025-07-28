package models

import "html/template"

// ToolCall represents a tool invocation
type ToolCall struct {
	ID          string
	Name        string
	Description string
	Input       template.HTML
	RawInput    interface{}       // Raw input data before formatting
	CompactView template.HTML     // Optional compact view for specific tools
	Result      *ProcessedEntry   // Tool result entry
	TaskEntries []*ProcessedEntry // For Task tool - sidechain entries
}