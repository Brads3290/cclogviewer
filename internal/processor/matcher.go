package processor

import (
	"github.com/Brads3290/cclogviewer/internal/models"
	"strings"
	"time"
)

// ToolCallMatcher handles matching tool calls with their results
type ToolCallMatcher struct {
	windowSize time.Duration
}

// NewToolCallMatcher creates a new tool call matcher
func NewToolCallMatcher() *ToolCallMatcher {
	return &ToolCallMatcher{
		windowSize: 5 * time.Minute, // Default 5 minute window
	}
}

// MatchToolCalls matches tool calls with their results
func (m *ToolCallMatcher) MatchToolCalls(state *ProcessingState) error {
	// Build maps for both main and sidechain tool calls
	mainToolCallMap := make(map[string]*models.ToolCall)
	sidechainToolCallMap := make(map[string]*models.ToolCall)

	// First, build tool call maps
	for _, entry := range state.Entries {
		if !entry.IsSidechain {
			for i := range entry.ToolCalls {
				mainToolCallMap[entry.ToolCalls[i].ID] = &entry.ToolCalls[i]
			}
		} else {
			for i := range entry.ToolCalls {
				sidechainToolCallMap[entry.ToolCalls[i].ID] = &entry.ToolCalls[i]
			}
		}
	}

	// Second, match tool results
	for _, entry := range state.Entries {
		if entry.IsToolResult && entry.ToolResultID != "" {
			var toolCall *models.ToolCall

			if !entry.IsSidechain {
				toolCall = mainToolCallMap[entry.ToolResultID]
			} else {
				toolCall = sidechainToolCallMap[entry.ToolResultID]
			}

			if toolCall != nil {
				toolCall.Result = entry
				// Check if the tool was interrupted
				if entry.IsError && strings.Contains(strings.ToLower(entry.Content), "request interrupted by user") {
					toolCall.IsInterrupted = true
				}
			}
		}
	}

	return nil
}

// FilterRootEntries filters entries to only include root conversation entries
func (m *ToolCallMatcher) FilterRootEntries(entries []*models.ProcessedEntry) []*models.ProcessedEntry {
	var rootEntries []*models.ProcessedEntry

	// Build a set of tool result IDs that have been matched
	matchedResults := make(map[string]bool)
	for _, entry := range entries {
		for _, toolCall := range entry.ToolCalls {
			if toolCall.Result != nil {
				matchedResults[toolCall.Result.UUID] = true
			}
		}
	}

	// Include only non-sidechain entries that aren't matched tool results
	for _, entry := range entries {
		if !entry.IsSidechain && !matchedResults[entry.UUID] {
			rootEntries = append(rootEntries, entry)
		}
	}

	return rootEntries
}

// findToolCall finds a tool call by ID in the state
func (m *ToolCallMatcher) findToolCall(state *ProcessingState, toolUseID string) *ToolCallContext {
	return state.ToolCallMap[toolUseID]
}

// isWithinWindow checks if two timestamps are within the matching window
func (m *ToolCallMatcher) isWithinWindow(callTime, resultTime time.Time) bool {
	if callTime.IsZero() || resultTime.IsZero() {
		return false
	}
	diff := resultTime.Sub(callTime)
	return diff >= 0 && diff <= m.windowSize
}
