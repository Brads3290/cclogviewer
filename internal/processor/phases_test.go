package processor

import (
	"testing"

	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/models"
)

func TestInitializeProcessingState(t *testing.T) {
	state := initializeProcessingState(10)

	if state == nil {
		t.Fatal("Expected non-nil state")
	}

	if state.ToolCallMap == nil {
		t.Error("Expected initialized ToolCallMap")
	}

	if state.ParentChildMap == nil {
		t.Error("Expected initialized ParentChildMap")
	}

	if cap(state.Entries) != 10 {
		t.Errorf("Expected entries capacity of 10, got %d", cap(state.Entries))
	}
}

func TestProcessAllEntries(t *testing.T) {
	entries := []models.LogEntry{
		{
			UUID:      "test-1",
			Type:      constants.TypeUser,
			Timestamp: "2024-01-01T10:00:00Z",
			Message:   []byte(`{"role":"user","content":"Hello"}`),
		},
		{
			UUID:      "test-2",
			Type:      constants.TypeAssistant,
			Timestamp: "2024-01-01T10:00:01Z",
			Message:   []byte(`{"role":"assistant","content":[{"type":"text","text":"Hi there!"}]}`),
		},
	}

	state := initializeProcessingState(len(entries))
	entryMap := make(map[string]*models.ProcessedEntry)

	processAllEntries(entries, state, entryMap)

	if len(state.Entries) != 2 {
		t.Errorf("Expected 2 processed entries, got %d", len(state.Entries))
	}

	if len(entryMap) != 2 {
		t.Errorf("Expected 2 entries in map, got %d", len(entryMap))
	}

	// Check first entry
	if entryMap["test-1"] == nil {
		t.Error("Expected entry test-1 in map")
	} else {
		entry := entryMap["test-1"]
		if entry.UUID != "test-1" {
			t.Errorf("Expected UUID test-1, got %s", entry.UUID)
		}
		if entry.Role != constants.RoleUser {
			t.Errorf("Expected role %s, got %s", constants.RoleUser, entry.Role)
		}
	}
}

func TestGetRootEntries(t *testing.T) {
	state := &ProcessingState{
		Entries: []*models.ProcessedEntry{
			{UUID: "1", IsSidechain: false},
			{UUID: "2", IsSidechain: true},
			{UUID: "3", IsSidechain: false},
			{UUID: "4", IsSidechain: true},
		},
	}

	rootEntries := getRootEntries(state)

	// The FilterRootEntries method filters out sidechain entries
	expectedCount := 2
	if len(rootEntries) != expectedCount {
		t.Errorf("Expected %d root entries, got %d", expectedCount, len(rootEntries))
	}
}

func TestCalculateAllTokens(t *testing.T) {
	entries := []*models.ProcessedEntry{
		{
			UUID: "1",
			TokenMetrics: models.TokenMetrics{
				InputTokens:     100,
				OutputTokens:    50,
				CacheReadTokens: 20,
			},
		},
		{
			UUID: "2",
			TokenMetrics: models.TokenMetrics{
				InputTokens:         200,
				CacheCreationTokens: 30,
			},
		},
	}

	calculateAllTokens(entries)

	// Check token calculations (excluding output tokens)
	if entries[0].TotalTokens != 120 { // 100 + 20 (excluding 50 output)
		t.Errorf("Expected TotalTokens=120 for entry 1, got %d", entries[0].TotalTokens)
	}

	if entries[1].TotalTokens != 230 { // 200 + 30 (no output tokens)
		t.Errorf("Expected TotalTokens=230 for entry 2, got %d", entries[1].TotalTokens)
	}
}
