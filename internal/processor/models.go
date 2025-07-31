package processor

import (
	"github.com/brads3290/cclogviewer/internal/models"
	"time"
)

// ProcessingState holds all state during entry processing.
type ProcessingState struct {
	Entries        []*models.ProcessedEntry
	ToolCallMap    map[string]*ToolCallContext
	ParentChildMap map[string][]string
	Index          int
}

// ToolCallContext tracks pending tool calls awaiting results.
type ToolCallContext struct {
	Entry      *models.ProcessedEntry
	ToolCall   *models.ToolCall
	CallTime   time.Time
}

// SidechainContext tracks Task tool sidechain conversations.
type SidechainContext struct {
	RootToolCallID string
	Entries        []*models.ProcessedEntry
}

