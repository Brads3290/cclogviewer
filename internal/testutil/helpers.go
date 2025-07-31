package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// GenerateTestUUID generates a UUID for testing
func GenerateTestUUID() string {
	return uuid.New().String()
}

// CreateTestLogEntry creates a LogEntry for testing
func CreateTestLogEntry(t *testing.T, entryType, content string) *models.LogEntry {
	t.Helper()

	message, err := json.Marshal(map[string]interface{}{
		"content": content,
	})
	require.NoError(t, err)

	return &models.LogEntry{
		UUID:      GenerateTestUUID(),
		Type:      entryType,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   message,
	}
}

// CreateTestProcessedEntry creates a ProcessedEntry for testing
func CreateTestProcessedEntry(t *testing.T, entryType, content string) *models.ProcessedEntry {
	t.Helper()

	logEntry := CreateTestLogEntry(t, entryType, content)
	return &models.ProcessedEntry{
		UUID:         logEntry.UUID,
		Type:         logEntry.Type,
		Timestamp:    logEntry.Timestamp,
		RawTimestamp: logEntry.Timestamp,
		Content:      content,
		Depth:        1,
	}
}

// CreateToolCallEntry creates a tool call entry for testing
func CreateToolCallEntry(t *testing.T, toolName string, input interface{}) *models.LogEntry {
	t.Helper()

	inputJSON, err := json.Marshal(input)
	require.NoError(t, err)

	message, err := json.Marshal(map[string]interface{}{
		"name":  toolName,
		"input": json.RawMessage(inputJSON),
	})
	require.NoError(t, err)

	return &models.LogEntry{
		UUID:      GenerateTestUUID(),
		Type:      "tool_call",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   message,
	}
}

// CreateToolResultEntry creates a tool result entry for testing
func CreateToolResultEntry(t *testing.T, toolUseID string, output string) *models.LogEntry {
	t.Helper()

	message, err := json.Marshal(map[string]interface{}{
		"output": output,
	})
	require.NoError(t, err)

	return &models.LogEntry{
		UUID:      GenerateTestUUID(),
		Type:      "tool_result",
		Timestamp: time.Now().Add(100 * time.Millisecond).Format(time.RFC3339),
		Message:   message,
		ToolUseResult: map[string]interface{}{
			"toolUseId": toolUseID,
		},
	}
}

// LoadTestFile loads a test fixture file
func LoadTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", path))
	require.NoError(t, err)
	return data
}

// AssertGoldenFile compares output with golden file
func AssertGoldenFile(t *testing.T, actual []byte, goldenPath string) {
	t.Helper()

	goldenFile := filepath.Join("testdata", goldenPath)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		err := os.WriteFile(goldenFile, actual, 0644)
		require.NoError(t, err)
		t.Log("Updated golden file:", goldenFile)
		return
	}

	expected, err := os.ReadFile(goldenFile)
	if os.IsNotExist(err) {
		err = os.WriteFile(goldenFile, actual, 0644)
		require.NoError(t, err)
		t.Log("Created golden file:", goldenFile)
		return
	}
	require.NoError(t, err)

	require.Equal(t, string(expected), string(actual), "Golden file mismatch")
}

// CreateTestToolCallWithResult creates a matched tool call and result pair
func CreateTestToolCallWithResult(t *testing.T, toolName string, delay time.Duration) (toolCall *models.LogEntry, toolResult *models.LogEntry) {
	t.Helper()

	toolCall = CreateToolCallEntry(t, toolName, map[string]string{"command": "test"})
	
	// Adjust timestamp for result
	resultTime, err := time.Parse(time.RFC3339, toolCall.Timestamp)
	require.NoError(t, err)
	
	toolResult = &models.LogEntry{
		UUID:      GenerateTestUUID(),
		Type:      "tool_result",
		Timestamp: resultTime.Add(delay).Format(time.RFC3339),
		Message:   json.RawMessage(`{"output": "test output"}`),
		ToolUseResult: map[string]interface{}{
			"toolUseId": toolCall.UUID,
		},
	}

	return toolCall, toolResult
}