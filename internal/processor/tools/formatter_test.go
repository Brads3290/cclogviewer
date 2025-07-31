package tools_test

import (
	"strings"
	"testing"

	"github.com/brads3290/cclogviewer/internal/processor/tools"
	"github.com/brads3290/cclogviewer/internal/processor/tools/formatters"
)

func TestFormatterRegistry(t *testing.T) {
	registry := tools.NewFormatterRegistry()

	// Register a test formatter
	editFormatter := formatters.NewEditFormatter()
	registry.Register(editFormatter)

	// Test formatting with valid data
	data := map[string]interface{}{
		"file_path":  "/test/file.go",
		"old_string": "hello",
		"new_string": "world",
	}

	html, err := registry.Format("Edit", data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if html == "" {
		t.Error("Expected non-empty HTML output")
	}

	// Test description
	desc := registry.GetDescription("Edit", data)
	if desc != "/test/file.go" {
		t.Errorf("Expected description '/test/file.go', got '%s'", desc)
	}

	// Test with replace_all
	data["replace_all"] = true
	desc = registry.GetDescription("Edit", data)
	if desc != "/test/file.go (replace all)" {
		t.Errorf("Expected description with replace all, got '%s'", desc)
	}
}

func TestEditFormatter(t *testing.T) {
	formatter := formatters.NewEditFormatter()

	// Test validation
	invalidData := map[string]interface{}{
		"file_path": "/test/file.go",
		// missing old_string and new_string
	}

	err := formatter.ValidateInput(invalidData)
	if err == nil {
		t.Error("Expected validation error for missing fields")
	}

	// Test with valid data
	validData := map[string]interface{}{
		"file_path":  "/test/file.go",
		"old_string": "line1\nline2\nline3",
		"new_string": "line1\nline2-modified\nline3",
	}

	err = formatter.ValidateInput(validData)
	if err != nil {
		t.Errorf("Expected no validation error, got %v", err)
	}

	html, err := formatter.FormatInput(validData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that output contains diff markers
	htmlStr := string(html)
	if !contains(htmlStr, "line-unchanged") {
		t.Error("Expected unchanged lines in diff")
	}
	if !contains(htmlStr, "line-removed") {
		t.Error("Expected removed lines in diff")
	}
	if !contains(htmlStr, "line-added") {
		t.Error("Expected added lines in diff")
	}
}

func TestMultiEditFormatter(t *testing.T) {
	formatter := formatters.NewMultiEditFormatter()

	// Test with multiple edits
	data := map[string]interface{}{
		"file_path": "/test/file.go",
		"edits": []interface{}{
			map[string]interface{}{
				"old_string": "hello",
				"new_string": "world",
			},
			map[string]interface{}{
				"old_string":  "foo",
				"new_string":  "bar",
				"replace_all": true,
			},
		},
	}

	html, err := formatter.FormatInput(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	htmlStr := string(html)
	if !contains(htmlStr, "Edit #1") {
		t.Error("Expected edit numbering")
	}
	if !contains(htmlStr, "Edit #2") {
		t.Error("Expected second edit")
	}
	if !contains(htmlStr, "(Replace All)") {
		t.Error("Expected replace all indicator")
	}
}

func TestBashFormatter(t *testing.T) {
	formatter := formatters.NewBashFormatter()

	// Test validation
	err := formatter.ValidateInput(map[string]interface{}{})
	if err == nil {
		t.Error("Expected validation error for missing command")
	}

	// Test with valid data
	data := map[string]interface{}{
		"command":     "ls -la",
		"description": "List files",
		"timeout":     5000.0,
	}

	err = formatter.ValidateInput(data)
	if err != nil {
		t.Errorf("Expected no validation error, got %v", err)
	}

	// Test FormatInputWithCWD
	html, err := formatter.FormatInputWithCWD(data, "/home/user")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	htmlStr := string(html)
	if !contains(htmlStr, "ls -la") {
		t.Error("Expected command in output")
	}
	if !contains(htmlStr, "/home/user") {
		t.Error("Expected CWD in output")
	}
	if !contains(htmlStr, "List files") {
		t.Error("Expected description in output")
	}
	if !contains(htmlStr, "5000ms") {
		t.Error("Expected timeout in output")
	}
}

func TestTodoWriteFormatter(t *testing.T) {
	formatter := formatters.NewTodoWriteFormatter()

	data := map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":  "Task 1",
				"status":   "completed",
				"priority": "high",
			},
			map[string]interface{}{
				"content":  "Task 2",
				"status":   "in_progress",
				"priority": "medium",
			},
			map[string]interface{}{
				"content":  "Task 3",
				"status":   "pending",
				"priority": "low",
			},
		},
	}

	html, err := formatter.FormatInput(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	htmlStr := string(html)
	if !contains(htmlStr, "Task 1") {
		t.Error("Expected task 1 in output")
	}
	if !contains(htmlStr, "✓") {
		t.Error("Expected completed icon")
	}
	if !contains(htmlStr, "⏳") {
		t.Error("Expected in progress icon")
	}
	if !contains(htmlStr, "○") {
		t.Error("Expected pending icon")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
