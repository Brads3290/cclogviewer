// +build integration

package cclogviewer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brads3290/cclogviewer/internal/parser"
	"github.com/brads3290/cclogviewer/internal/processor"
	"github.com/brads3290/cclogviewer/internal/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConvertFile simulates the full conversion pipeline
func ConvertFile(inputPath, outputPath string, debugMode bool) error {
	// Parse JSONL file
	entries, err := parser.ReadJSONLFile(inputPath)
	if err != nil {
		return err
	}

	// Process entries
	processedEntries := processor.ProcessEntries(entries)

	// Generate HTML
	return renderer.GenerateHTML(processedEntries, outputPath, debugMode)
}

func TestEndToEnd_SimpleConversion(t *testing.T) {
	// Read test JSONL
	inputPath := "testdata/fixtures/valid/simple.jsonl"
	outputPath := filepath.Join(t.TempDir(), "output.html")

	// Run full pipeline
	err := ConvertFile(inputPath, outputPath, false)
	require.NoError(t, err)

	// Verify output exists
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(content)
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Hello")
	assert.Contains(t, html, "Hi there! How can I help you today?")
}

func TestEndToEnd_WithToolCalls(t *testing.T) {
	inputPath := "testdata/fixtures/valid/with_tools.jsonl"
	outputPath := filepath.Join(t.TempDir(), "with_tools.html")

	err := ConvertFile(inputPath, outputPath, false)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(content)
	// Verify tool call is rendered
	assert.Contains(t, html, "Bash")
	assert.Contains(t, html, "ls -la")
	// Verify tool result is rendered
	assert.Contains(t, html, "file1.txt")
	assert.Contains(t, html, "file2.txt")
}

func TestEndToEnd_WithSidechain(t *testing.T) {
	inputPath := "testdata/fixtures/valid/with_sidechain.jsonl"
	outputPath := filepath.Join(t.TempDir(), "with_sidechain.html")

	err := ConvertFile(inputPath, outputPath, false)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(content)
	// Verify Task tool is rendered
	assert.Contains(t, html, "Task")
	// Verify sidechain content is included
	assert.Contains(t, html, "Analyze codebase")
	assert.Contains(t, html, "standard Go project layout")
}

func TestEndToEnd_ErrorHandling(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		err := ConvertFile("does_not_exist.jsonl", "output.html", false)
		assert.Error(t, err)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		inputPath := "testdata/fixtures/invalid/malformed.jsonl"
		outputPath := filepath.Join(t.TempDir(), "malformed.html")

		// Should handle gracefully - skip bad lines
		err := ConvertFile(inputPath, outputPath, false)
		require.NoError(t, err)

		// Should still have valid entries
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Valid line")
		assert.Contains(t, string(content), "Another valid line")
	})

	t.Run("empty file", func(t *testing.T) {
		inputPath := "testdata/fixtures/invalid/empty.jsonl"
		outputPath := filepath.Join(t.TempDir(), "empty.html")

		err := ConvertFile(inputPath, outputPath, false)
		require.NoError(t, err)

		// Should generate valid HTML even with no entries
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "<!DOCTYPE html>")
	})
}

func TestEndToEnd_DebugMode(t *testing.T) {
	inputPath := "testdata/fixtures/valid/simple.jsonl"
	outputPath := filepath.Join(t.TempDir(), "debug.html")

	// Run with debug mode enabled
	err := ConvertFile(inputPath, outputPath, true)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Should include debug logging
	html := string(content)
	assert.Contains(t, html, "debugLog")
	assert.Contains(t, html, "[DEBUG]")
}

func TestEndToEnd_UnicodeHandling(t *testing.T) {
	inputPath := "testdata/fixtures/edge_cases/unicode.jsonl"
	outputPath := filepath.Join(t.TempDir(), "unicode.html")

	err := ConvertFile(inputPath, outputPath, false)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(content)
	// Verify Unicode is preserved
	assert.Contains(t, html, "🚀")
	assert.Contains(t, html, "🎉")
	assert.Contains(t, html, "αβγδε")
}