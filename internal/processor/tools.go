package processor

import (
	"cclogviewer/internal/models"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strings"
)

// ProcessToolUse processes a tool use message and returns a ToolCall
func ProcessToolUse(toolUse map[string]interface{}) models.ToolCall {
	tool := models.ToolCall{
		ID:   GetStringValue(toolUse, "id"),
		Name: GetStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.Description = GetStringValue(input, "description")
		
		// Special handling for Edit and MultiEdit tools
		if tool.Name == "Edit" {
			tool.Input = formatEditToolInput(input)
		} else if tool.Name == "MultiEdit" {
			tool.Input = formatMultiEditToolInput(input)
		} else {
			// Format the input as JSON
			inputJSON, _ := json.MarshalIndent(input, "", "  ")
			tool.Input = template.HTML(fmt.Sprintf(`<pre class="tool-input">%s</pre>`, html.EscapeString(string(inputJSON))))
		}
	}

	return tool
}

// diffLine represents a line in the diff with its type and content
type diffLine struct {
	Type    string // "unchanged", "removed", "added"
	Content string
	LineNum int
}

// computeLineDiff computes a simple line-based diff between two strings
func computeLineDiff(oldStr, newStr string) []diffLine {
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")
	
	// For a simple implementation, we'll use a basic LCS approach
	// This is not as sophisticated as a full diff algorithm but works well for most cases
	diff := []diffLine{}
	
	// If strings are identical, return all unchanged lines
	if oldStr == newStr {
		for i, line := range oldLines {
			diff = append(diff, diffLine{
				Type:    "unchanged",
				Content: line,
				LineNum: i + 1,
			})
		}
		return diff
	}
	
	// Simple diff: find longest common subsequence
	lcs := longestCommonSubsequence(oldLines, newLines)
	
	// Build diff from LCS
	oldIdx, newIdx := 0, 0
	lcsIdx := 0
	lineNum := 1
	
	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if lcsIdx < len(lcs) && oldIdx < len(oldLines) && newIdx < len(newLines) &&
			oldLines[oldIdx] == lcs[lcsIdx] && newLines[newIdx] == lcs[lcsIdx] {
			// Common line
			diff = append(diff, diffLine{
				Type:    "unchanged",
				Content: oldLines[oldIdx],
				LineNum: lineNum,
			})
			oldIdx++
			newIdx++
			lcsIdx++
			lineNum++
		} else if oldIdx < len(oldLines) && (lcsIdx >= len(lcs) || oldLines[oldIdx] != lcs[lcsIdx]) {
			// Removed line
			diff = append(diff, diffLine{
				Type:    "removed",
				Content: oldLines[oldIdx],
				LineNum: lineNum,
			})
			oldIdx++
			lineNum++
		} else if newIdx < len(newLines) && (lcsIdx >= len(lcs) || newLines[newIdx] != lcs[lcsIdx]) {
			// Added line
			diff = append(diff, diffLine{
				Type:    "added",
				Content: newLines[newIdx],
				LineNum: lineNum,
			})
			newIdx++
			lineNum++
		}
	}
	
	return diff
}

// longestCommonSubsequence finds the LCS of two string slices
func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	
	// Build the DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	
	// Reconstruct the LCS
	lcs := []string{}
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	
	return lcs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// formatEditToolInput creates a diff view for Edit tool inputs
func formatEditToolInput(input map[string]interface{}) template.HTML {
	filePath := GetStringValue(input, "file_path")
	oldString := GetStringValue(input, "old_string")
	newString := GetStringValue(input, "new_string")
	replaceAll := false
	if val, ok := input["replace_all"].(bool); ok {
		replaceAll = val
	}

	// Compute the unified diff
	diffLines := computeLineDiff(oldString, newString)

	// Build the diff HTML
	var diffHTML strings.Builder
	diffHTML.WriteString(`<div class="edit-diff">`)
	diffHTML.WriteString(fmt.Sprintf(`<div class="diff-header">📝 Edit: <span class="file-path">%s</span>`, html.EscapeString(filePath)))
	if replaceAll {
		diffHTML.WriteString(` <span class="replace-all">(Replace All)</span>`)
	}
	diffHTML.WriteString(`</div>`)
	diffHTML.WriteString(`<div class="diff-content unified">`)
	diffHTML.WriteString(`<pre class="diff-code">`)
	
	// Display the unified diff
	for i, line := range diffLines {
		var lineClass string
		var prefix string
		
		switch line.Type {
		case "removed":
			lineClass = "line-removed"
			prefix = "-"
		case "added":
			lineClass = "line-added"
			prefix = "+"
		default:
			lineClass = "line-unchanged"
			prefix = " "
		}
		
		diffHTML.WriteString(fmt.Sprintf(`<span class="diff-line %s"><span class="line-number">%3d</span><span class="line-prefix">%s</span> %s</span>`,
			lineClass, line.LineNum, prefix, html.EscapeString(line.Content)))
		
		if i < len(diffLines)-1 {
			diffHTML.WriteString("\n")
		}
	}
	
	diffHTML.WriteString(`</pre>`)
	diffHTML.WriteString(`</div></div>`)
	return template.HTML(diffHTML.String())
}

// formatMultiEditToolInput creates a diff view for MultiEdit tool inputs
func formatMultiEditToolInput(input map[string]interface{}) template.HTML {
	filePath := GetStringValue(input, "file_path")
	edits, ok := input["edits"].([]interface{})
	if !ok {
		// Fallback to JSON display
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		return template.HTML(fmt.Sprintf(`<pre class="tool-input">%s</pre>`, html.EscapeString(string(inputJSON))))
	}

	// Build the multi-edit HTML
	var multiEditHTML strings.Builder
	multiEditHTML.WriteString(`<div class="multi-edit">`)
	multiEditHTML.WriteString(fmt.Sprintf(`<div class="diff-header">📝 MultiEdit: <span class="file-path">%s</span> <span class="replace-all">%d edits</span></div>`, 
		html.EscapeString(filePath), len(edits)))
	
	// Process each edit
	for i, editInterface := range edits {
		edit, ok := editInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		oldString := GetStringValue(edit, "old_string")
		newString := GetStringValue(edit, "new_string")
		replaceAll := false
		if val, ok := edit["replace_all"].(bool); ok {
			replaceAll = val
		}
		
		// Compute the unified diff for this edit
		diffLines := computeLineDiff(oldString, newString)
		
		if i == 0 {
			multiEditHTML.WriteString(`<div class="edit-item" style="padding-top: 0; margin-top: 0;">`)
		} else {
			multiEditHTML.WriteString(fmt.Sprintf(`<div class="edit-item" style="border-top: 1px solid #dee2e6; padding-top: 10px; margin-top: 10px;">`))
		}
		
		multiEditHTML.WriteString(fmt.Sprintf(`<div style="color: #6c757d; font-size: 0.85em; margin-bottom: 5px;">Edit #%d`, i+1))
		if replaceAll {
			multiEditHTML.WriteString(` <span class="replace-all">(Replace All)</span>`)
		}
		multiEditHTML.WriteString(`</div>`)
		
		multiEditHTML.WriteString(`<div class="diff-content unified">`)
		multiEditHTML.WriteString(`<pre class="diff-code">`)
		
		// Display the unified diff
		for j, line := range diffLines {
			var lineClass string
			var prefix string
			
			switch line.Type {
			case "removed":
				lineClass = "line-removed"
				prefix = "-"
			case "added":
				lineClass = "line-added"
				prefix = "+"
			default:
				lineClass = "line-unchanged"
				prefix = " "
			}
			
			multiEditHTML.WriteString(fmt.Sprintf(`<span class="diff-line %s"><span class="line-number">%3d</span><span class="line-prefix">%s</span> %s</span>`,
				lineClass, line.LineNum, prefix, html.EscapeString(line.Content)))
			
			if j < len(diffLines)-1 {
				multiEditHTML.WriteString("\n")
			}
		}
		
		multiEditHTML.WriteString(`</pre>`)
		multiEditHTML.WriteString(`</div>`)
		multiEditHTML.WriteString(`</div>`)
	}
	
	multiEditHTML.WriteString(`</div>`)
	return template.HTML(multiEditHTML.String())
}