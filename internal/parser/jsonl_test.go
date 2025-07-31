package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJSONLFile(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid simple file",
			file:      "fixtures/valid/simple.jsonl",
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "file with tool calls",
			file:      "fixtures/valid/with_tools.jsonl",
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "file with sidechain",
			file:      "fixtures/valid/with_sidechain.jsonl",
			wantCount: 8,
			wantErr:   false,
		},
		{
			name:      "empty file",
			file:      "fixtures/invalid/empty.jsonl",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "malformed JSON",
			file:      "fixtures/invalid/malformed.jsonl",
			wantCount: 3, // Should skip bad lines
			wantErr:   false,
		},
		{
			name:    "non-existent file",
			file:    "does_not_exist.jsonl",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the test directory path
			testDir := getTestDataDir(t)
			path := filepath.Join(testDir, tt.file)
			
			entries, err := ReadJSONLFile(path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, entries, tt.wantCount)

			// Verify entries have required fields
			for _, entry := range entries {
				assert.NotEmpty(t, entry.UUID)
				assert.NotEmpty(t, entry.Type)
				assert.NotEmpty(t, entry.Timestamp)
			}
		})
	}
}

func TestReadJSONLFile_SpecialCharacters(t *testing.T) {
	testDir := getTestDataDir(t)
	path := filepath.Join(testDir, "fixtures/edge_cases/unicode.jsonl")
	
	entries, err := ReadJSONLFile(path)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify Unicode content is preserved
	for _, entry := range entries {
		content := string(entry.Message)
		assert.True(t, strings.Contains(content, "🚀") || strings.Contains(content, "αβγδε"),
			"Unicode characters should be preserved")
	}
}

func TestReadJSONLFile_SkipsSummaryMessages(t *testing.T) {
	// Create a temporary file with summary messages
	tmpfile, err := os.CreateTemp("", "test_summary_*.jsonl")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := `{"uuid":"msg-001","type":"message","timestamp":"2024-01-01T10:00:00Z","message":{"content":"Regular message"}}
{"uuid":"sum-001","type":"summary","timestamp":"2024-01-01T10:00:01Z","message":{"content":"Summary message"}}
{"uuid":"msg-002","type":"message","timestamp":"2024-01-01T10:00:02Z","message":{"content":"Another regular message"}}`

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	entries, err := ReadJSONLFile(tmpfile.Name())
	require.NoError(t, err)
	assert.Len(t, entries, 2, "Should skip summary messages")

	// Verify no summary messages in result
	for _, entry := range entries {
		assert.NotEqual(t, "summary", entry.Type)
	}
}

func TestReadJSONLFile_LargeLines(t *testing.T) {
	// Create a temporary file with a very large line
	tmpfile, err := os.CreateTemp("", "test_large_*.jsonl")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create a large content string
	largeContent := strings.Repeat("a", 100000)
	line := `{"uuid":"msg-001","type":"message","timestamp":"2024-01-01T10:00:00Z","message":{"content":"` + largeContent + `"}}`
	
	_, err = tmpfile.WriteString(line + "\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	entries, err := ReadJSONLFile(tmpfile.Name())
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	
	// Verify the large content was read correctly
	message := string(entries[0].Message)
	assert.Contains(t, message, largeContent)
}

func BenchmarkReadJSONLFile(b *testing.B) {
	// Create a benchmark file with many entries
	tmpfile, err := os.CreateTemp("", "bench_*.jsonl")
	require.NoError(b, err)
	defer os.Remove(tmpfile.Name())

	// Write 1000 entries
	for i := 0; i < 1000; i++ {
		line := `{"uuid":"msg-%d","type":"message","timestamp":"2024-01-01T10:00:00Z","message":{"content":"Test message %d"}}`
		_, err = tmpfile.WriteString(strings.ReplaceAll(line, "%d", string(rune(i))) + "\n")
		require.NoError(b, err)
	}
	require.NoError(b, tmpfile.Close())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ReadJSONLFile(tmpfile.Name())
	}
}

// getTestDataDir returns the path to the testdata directory
func getTestDataDir(t *testing.T) string {
	t.Helper()
	
	// Try to find the testdata directory relative to the test file
	wd, err := os.Getwd()
	require.NoError(t, err)
	
	// Go up to the project root
	for {
		if _, err := os.Stat(filepath.Join(wd, "testdata")); err == nil {
			return filepath.Join(wd, "testdata")
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	
	t.Fatal("Could not find testdata directory")
	return ""
}