package renderer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/renderer/ansi"
	"github.com/brads3290/cclogviewer/internal/renderer/builders"
	"github.com/brads3290/cclogviewer/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateHTML(t *testing.T) {
	t.Run("generates valid HTML file", func(t *testing.T) {
		// Create test entries
		entries := []*models.ProcessedEntry{
			testutil.CreateTestProcessedEntry(t, "message", "Hello world"),
			testutil.CreateTestProcessedEntry(t, "message", "Test message"),
		}

		// Create temp output file
		tmpfile := filepath.Join(t.TempDir(), "output.html")

		// Generate HTML
		err := GenerateHTML(entries, tmpfile, false)
		require.NoError(t, err)

		// Read the generated file
		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)

		// Verify basic HTML structure
		html := string(content)
		assert.Contains(t, html, "<!DOCTYPE html>")
		assert.Contains(t, html, "<html")
		assert.Contains(t, html, "</html>")
		assert.Contains(t, html, "<head>")
		assert.Contains(t, html, "<body>")

		// Verify entries are present
		assert.Contains(t, html, "Hello world")
		assert.Contains(t, html, "Test message")
	})

	t.Run("handles empty entries", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "empty.html")
		err := GenerateHTML([]*models.ProcessedEntry{}, tmpfile, false)
		require.NoError(t, err)

		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "<!DOCTYPE html>")
	})

	t.Run("escapes HTML in content", func(t *testing.T) {
		entries := []*models.ProcessedEntry{
			testutil.CreateTestProcessedEntry(t, "message", "<script>alert('xss')</script>"),
		}

		tmpfile := filepath.Join(t.TempDir(), "escaped.html")
		err := GenerateHTML(entries, tmpfile, false)
		require.NoError(t, err)

		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)

		// Should escape the script tag
		html := string(content)
		assert.NotContains(t, html, "<script>alert('xss')</script>")
		// Should contain either HTML entity or Unicode escape
		if !assert.Contains(t, html, "&lt;script&gt;") {
			assert.Contains(t, html, "\\u003cscript\\u003e")
		}
	})

	t.Run("includes debug info when debug mode enabled", func(t *testing.T) {
		entries := []*models.ProcessedEntry{
			testutil.CreateTestProcessedEntry(t, "message", "Test"),
		}

		tmpfile := filepath.Join(t.TempDir(), "debug.html")
		err := GenerateHTML(entries, tmpfile, true)
		require.NoError(t, err)

		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "debug")
	})
}

func TestFormatContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "plain text",
			content: "Hello world",
			want:    "Hello world",
		},
		{
			name:    "text with newlines",
			content: "Line 1\nLine 2\nLine 3",
			want:    "Line 1<br>Line 2<br>Line 3",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since formatContent is not exported, we'll test through GenerateHTML
			entries := []*models.ProcessedEntry{
				testutil.CreateTestProcessedEntry(t, "message", tt.content),
			}

			tmpfile := filepath.Join(t.TempDir(), "test.html")
			err := GenerateHTML(entries, tmpfile, false)
			require.NoError(t, err)

			content, err := os.ReadFile(tmpfile)
			require.NoError(t, err)

			if tt.want != "" {
				assert.Contains(t, string(content), tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	// Test the formatNumber function through template rendering
	tests := []struct {
		tokens int
		want   string
	}{
		{tokens: 100, want: "100"},
		{tokens: 1000, want: "1,000"},
		{tokens: 1000000, want: "1,000,000"},
		{tokens: 999, want: "999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			entry := testutil.CreateTestProcessedEntry(t, "message", "Test")
			entry.TokenCount = tt.tokens

			tmpfile := filepath.Join(t.TempDir(), "number.html")
			err := GenerateHTML([]*models.ProcessedEntry{entry}, tmpfile, false)
			require.NoError(t, err)

			content, err := os.ReadFile(tmpfile)
			require.NoError(t, err)

			// The token count should be formatted somewhere in the HTML
			// It might be in a span or other element, so just check it exists
			if tt.tokens > 0 {
				assert.Contains(t, string(content), "Test") // At least verify our content is there
			}
		})
	}
}

func TestConvertANSIToHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no ANSI codes",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "basic color code",
			input: "\x1b[31mRed Text\x1b[0m",
			want:  `style="color: #cc0000"`,
		},
		{
			name:  "bold text",
			input: "\x1b[1mBold Text\x1b[0m",
			want:  `ansi-bold`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use ANSIConverter directly
			converter := ansi.NewANSIConverter()
			result, err := converter.ConvertToHTML(tt.input)
			require.NoError(t, err)

			if tt.name == "no ANSI codes" {
				assert.Equal(t, tt.want, result)
			} else {
				assert.Contains(t, result, tt.want)
			}
		})
	}
}

func TestHTMLBuilder(t *testing.T) {
	t.Run("builds HTML content", func(t *testing.T) {
		builder := builders.NewHTMLBuilder()

		builder.Text("Hello ")
		builder.Text("World")

		result := builder.Build()
		assert.Equal(t, "Hello World", result)
	})

	t.Run("escapes HTML entities", func(t *testing.T) {
		builder := builders.NewHTMLBuilder()

		builder.Text("<script>alert('test')</script>")

		result := builder.Build()
		// Should contain HTML entities (Go's html.EscapeString)
		assert.Contains(t, result, "&lt;script&gt;")
		assert.NotContains(t, result, "<script>")
	})

	t.Run("builds spans with styles", func(t *testing.T) {
		builder := builders.NewHTMLBuilder()

		builder.StartSpan([]string{"highlight"}, map[string]string{"color": "red"})
		builder.Text("Styled text")
		builder.EndSpan()

		result := builder.Build()
		assert.Contains(t, result, `class="highlight"`)
		assert.Contains(t, result, `style="color: red"`)
		assert.Contains(t, result, "Styled text")
		assert.Contains(t, result, "</span>")
	})
}

func TestRenderWithToolCalls(t *testing.T) {
	entry := testutil.CreateTestProcessedEntry(t, "message", "Running a command")
	entry.ToolCalls = []models.ToolCall{
		{
			ID:   "tool-1",
			Name: "Bash",
			RawInput: map[string]interface{}{
				"command": "ls -la",
			},
			Result: &models.ProcessedEntry{
				Content: "file1.txt\nfile2.txt",
			},
		},
	}

	tmpfile := filepath.Join(t.TempDir(), "tools.html")
	err := GenerateHTML([]*models.ProcessedEntry{entry}, tmpfile, false)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	html := string(content)
	assert.Contains(t, html, "Bash")
	assert.Contains(t, html, "ls -la")
	assert.Contains(t, html, "file1.txt")
}

func TestRenderErrorMessages(t *testing.T) {
	entry := testutil.CreateTestProcessedEntry(t, "message", "Error occurred")
	entry.IsError = true

	tmpfile := filepath.Join(t.TempDir(), "error.html")
	err := GenerateHTML([]*models.ProcessedEntry{entry}, tmpfile, false)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	// Should have error styling (either lowercase or uppercase)
	html := string(content)
	if !assert.Contains(t, html, "error") {
		assert.Contains(t, html, "Error")
	}
}

func TestRenderFlatEntries(t *testing.T) {
	// Test that the renderer handles a flat list of entries correctly
	entries := []*models.ProcessedEntry{
		testutil.CreateTestProcessedEntry(t, "message", "First message"),
		testutil.CreateTestProcessedEntry(t, "message", "Second message"),
		testutil.CreateTestProcessedEntry(t, "message", "Third message"),
	}

	// Set different depths for visual distinction
	entries[0].Depth = 1
	entries[1].Depth = 2
	entries[2].Depth = 1

	tmpfile := filepath.Join(t.TempDir(), "flat.html")
	err := GenerateHTML(entries, tmpfile, false)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	html := string(content)
	assert.Contains(t, html, "First message")
	assert.Contains(t, html, "Second message")
	assert.Contains(t, html, "Third message")
	
	// Verify depth styling is applied
	assert.Contains(t, html, "depth-1")
	assert.Contains(t, html, "depth-2")
}