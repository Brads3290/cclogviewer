package processor

import (
	"testing"

	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/models"
)

func TestMatchToolCalls(t *testing.T) {
	// Create test entries with tool calls and results
	entries := []*models.ProcessedEntry{
		{
			UUID: "msg-1",
			Role: constants.RoleAssistant,
			ToolCalls: []models.ToolCall{
				{ID: "tool-1", Name: constants.ToolNameBash},
				{ID: "tool-2", Name: constants.ToolNameEdit},
			},
		},
		{
			UUID:         "result-1",
			Role:         constants.RoleUser,
			IsToolResult: true,
			ToolResultID: "tool-1",
			Content:      "Command output",
		},
		{
			UUID:         "result-2",
			Role:         constants.RoleUser,
			IsToolResult: true,
			ToolResultID: "tool-2",
			Content:      "Edit successful",
		},
	}

	matcher := NewToolCallMatcher()
	err := matcher.MatchToolCalls(entries)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify tool calls were matched with results
	toolCall1 := &entries[0].ToolCalls[0]
	if toolCall1.Result == nil {
		t.Error("Expected tool-1 to have a result")
	} else if toolCall1.Result.UUID != "result-1" {
		t.Errorf("Expected tool-1 result UUID to be result-1, got %s", toolCall1.Result.UUID)
	}

	toolCall2 := &entries[0].ToolCalls[1]
	if toolCall2.Result == nil {
		t.Error("Expected tool-2 to have a result")
	} else if toolCall2.Result.UUID != "result-2" {
		t.Errorf("Expected tool-2 result UUID to be result-2, got %s", toolCall2.Result.UUID)
	}
}

func TestFilterRootEntries(t *testing.T) {
	entries := []*models.ProcessedEntry{
		{UUID: "1", IsSidechain: false},
		{UUID: "2", IsSidechain: true},
		{UUID: "3", IsSidechain: false, IsToolResult: true, ToolResultID: "tool-1"},
		{UUID: "4", IsSidechain: false, ToolCalls: []models.ToolCall{{ID: "tool-1", Result: &models.ProcessedEntry{UUID: "3"}}}},
	}

	matcher := NewToolCallMatcher()
	rootEntries := matcher.FilterRootEntries(entries)

	// Should filter out:
	// - Entry 2 (sidechain)
	// - Entry 3 (matched tool result)
	expectedCount := 2
	if len(rootEntries) != expectedCount {
		t.Errorf("Expected %d root entries, got %d", expectedCount, len(rootEntries))
	}

	// Verify correct entries were included
	foundEntry1 := false
	foundEntry4 := false
	for _, entry := range rootEntries {
		if entry.UUID == "1" {
			foundEntry1 = true
		}
		if entry.UUID == "4" {
			foundEntry4 = true
		}
	}

	if !foundEntry1 {
		t.Error("Expected entry 1 to be in root entries")
	}
	if !foundEntry4 {
		t.Error("Expected entry 4 to be in root entries")
	}
}
