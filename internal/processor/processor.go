package processor

import (
	"fmt"
	"github.com/Brads3290/cclogviewer/internal/models"
	"time"
)

// EntryProcessor defines the interface for processing log entries
type EntryProcessor interface {
	CanProcess(entry *models.LogEntry) bool
	Process(entry *models.LogEntry, state *ProcessingState) (*models.ProcessedEntry, error)
}

// ProcessorChain manages a chain of processors
type ProcessorChain struct {
	processors []EntryProcessor
}

// NewProcessorChain creates a new processor chain with default processors
func NewProcessorChain() *ProcessorChain {
	return &ProcessorChain{
		processors: []EntryProcessor{
			&ToolCallProcessor{},
			&ToolResultProcessor{},
			&MessageProcessor{},
		},
	}
}

// Process runs the entry through the chain of processors
func (pc *ProcessorChain) Process(entry *models.LogEntry, state *ProcessingState) (*models.ProcessedEntry, error) {
	for _, processor := range pc.processors {
		if processor.CanProcess(entry) {
			return processor.Process(entry, state)
		}
	}
	return nil, fmt.Errorf("no processor found for entry type: %s", entry.Type)
}

// ToolCallProcessor handles entries with tool calls
type ToolCallProcessor struct{}

// CanProcess checks if this processor can handle the entry
func (tcp *ToolCallProcessor) CanProcess(entry *models.LogEntry) bool {
	return entry.Type == "assistant"
}

// Process processes the entry
func (tcp *ToolCallProcessor) Process(entry *models.LogEntry, state *ProcessingState) (*models.ProcessedEntry, error) {
	processed := processEntry(*entry)

	// Track tool calls for later matching
	for i := range processed.ToolCalls {
		toolCall := &processed.ToolCalls[i]
		state.ToolCallMap[toolCall.ID] = &ToolCallContext{
			Entry:    processed,
			ToolCall: toolCall,
			CallTime: parseTimestamp(entry.Timestamp),
		}
	}

	return processed, nil
}

// ToolResultProcessor handles tool result entries
type ToolResultProcessor struct{}

// CanProcess checks if this processor can handle the entry
func (trp *ToolResultProcessor) CanProcess(entry *models.LogEntry) bool {
	return entry.Type == "user" && entry.ToolUseResult != nil
}

// Process processes the entry
func (trp *ToolResultProcessor) Process(entry *models.LogEntry, state *ProcessingState) (*models.ProcessedEntry, error) {
	processed := processEntry(*entry)

	// This will be matched later in the matching phase
	return processed, nil
}

// MessageProcessor handles regular message entries
type MessageProcessor struct{}

// CanProcess checks if this processor can handle the entry
func (mp *MessageProcessor) CanProcess(entry *models.LogEntry) bool {
	return entry.Type == "user" || entry.Type == "assistant"
}

// Process processes the entry
func (mp *MessageProcessor) Process(entry *models.LogEntry, state *ProcessingState) (*models.ProcessedEntry, error) {
	return processEntry(*entry), nil
}

// parseTimestamp parses a timestamp string into a time.Time
func parseTimestamp(ts string) time.Time {
	t, _ := time.Parse(time.RFC3339, ts)
	return t
}
