package processor

import (
	"github.com/brads3290/cclogviewer/internal/models"
	"time"
)

// ProcessingState holds all state during entry processing
type ProcessingState struct {
	Entries        []*models.ProcessedEntry
	ToolCallMap    map[string]*ToolCallContext
	ParentChildMap map[string][]string
	RootParent     string
	Index          int
}

// ToolCallContext tracks pending tool calls
type ToolCallContext struct {
	Entry      *models.ProcessedEntry
	ToolCall   *models.ToolCall
	CallTime   time.Time
	ParentID   string
	IsComplete bool
}

// SidechainContext tracks active sidechain conversations
type SidechainContext struct {
	RootToolCallID string
	StartIndex     int
	EndIndex       int
	Entries        []*models.ProcessedEntry
}

// MatchingOptions controls how tool calls are matched
type MatchingOptions struct {
	WindowSize time.Duration
}
