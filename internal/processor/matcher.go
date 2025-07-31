package processor

import (
	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/models"
	"strings"
)

// ToolCallMatcher matches tool calls with their results.
type ToolCallMatcher struct {
}

// NewToolCallMatcher creates a new tool call matcher
func NewToolCallMatcher() *ToolCallMatcher {
	return &ToolCallMatcher{}
}

// MatchToolCalls uses a 5-minute window to prevent false matches in long conversations.
func (m *ToolCallMatcher) MatchToolCalls(entries []*models.ProcessedEntry) error {
	// Build maps for both main and sidechain tool calls
	mainToolCallMap := make(map[string]*models.ToolCall)
	sidechainToolCallMap := make(map[string]*models.ToolCall)

	// First, build tool call maps
	for _, entry := range entries {
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
	for _, entry := range entries {
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
				if entry.IsError && strings.Contains(strings.ToLower(entry.Content), constants.UserInterruptionPattern) {
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

