package processor

import (
	"testing"

	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessEntries_Simple(t *testing.T) {
	tests := []struct {
		name          string
		entries       []models.LogEntry
		wantRootCount int
	}{
		{
			name: "simple conversation",
			entries: []models.LogEntry{
				*testutil.CreateTestLogEntry(t, "message", "Hello"),
				*testutil.CreateTestLogEntry(t, "message", "Hi there"),
			},
			wantRootCount: 2,
		},
		{
			name:          "empty entries",
			entries:       []models.LogEntry{},
			wantRootCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessEntries(tt.entries)
			assert.Len(t, result, tt.wantRootCount)
		})
	}
}

func TestFormatTimestamp_Simple(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid RFC3339 timestamp",
			input: "2024-01-01T15:30:45Z",
			want:  "15:30:45",
		},
		{
			name:  "invalid timestamp",
			input: "invalid",
			want:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimestamp(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeText_Simple(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes extra spaces",
			input: "hello   world",
			want:  "hello world",
		},
		{
			name:  "removes newlines",
			input: "hello\nworld",
			want:  "hello world",
		},
		{
			name:  "trims whitespace",
			input: "  hello world  ",
			want:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeText(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractXMLContent_Simple(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		tag     string
		want    string
	}{
		{
			name: "extracts content",
			text: "prefix <tag>content</tag> suffix",
			tag:  "tag",
			want: "content",
		},
		{
			name: "missing start tag",
			text: "no start tag here</tag>",
			tag:  "tag",
			want: "",
		},
		{
			name: "missing end tag",
			text: "<tag>no end tag here",
			tag:  "tag",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractXMLContent(tt.text, tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckMissingToolResults_Simple(t *testing.T) {
	t.Run("marks missing tool result", func(t *testing.T) {
		entry := &models.ProcessedEntry{
			ToolCalls: []models.ToolCall{
				{Name: "Bash", Result: nil},
			},
		}

		checkMissingToolResults(entry)
		assert.True(t, entry.ToolCalls[0].HasMissingResult)
	})

	t.Run("does not mark when result exists", func(t *testing.T) {
		entry := &models.ProcessedEntry{
			ToolCalls: []models.ToolCall{
				{Name: "Bash", Result: &models.ProcessedEntry{}},
			},
		}

		checkMissingToolResults(entry)
		assert.False(t, entry.ToolCalls[0].HasMissingResult)
	})
}

func TestCalculateTokensForEntry_Simple(t *testing.T) {
	t.Run("calculates total tokens excluding output", func(t *testing.T) {
		entry := &models.ProcessedEntry{
			TokenMetrics: models.TokenMetrics{
				InputTokens:         100,
				OutputTokens:        200,
				CacheReadTokens:     50,
				CacheCreationTokens: 25,
			},
		}

		calculateTokensForEntry(entry)
		assert.Equal(t, 175, entry.TotalTokens)
	})
}

func TestProcessEntry_Simple(t *testing.T) {
	entry := testutil.CreateTestLogEntry(t, "message", "Test content")
	result := processEntry(*entry)
	
	require.NotNil(t, result)
	assert.Equal(t, entry.UUID, result.UUID)
	assert.Equal(t, entry.Type, result.Type)
	assert.NotEmpty(t, result.Timestamp)
}