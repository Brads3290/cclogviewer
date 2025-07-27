package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogEntry represents a single line from the JSONL file
type LogEntry struct {
	ParentUUID   *string         `json:"parentUuid"`
	IsSidechain  bool            `json:"isSidechain"`
	UserType     string          `json:"userType"`
	CWD          string          `json:"cwd"`
	SessionID    string          `json:"sessionId"`
	Version      string          `json:"version"`
	GitBranch    string          `json:"gitBranch"`
	Type         string          `json:"type"`
	Message      json.RawMessage `json:"message"`
	RequestID    string          `json:"requestId"`
	UUID         string          `json:"uuid"`
	Timestamp    string          `json:"timestamp"`
	IsMeta       bool            `json:"isMeta"`
	ToolUseResult interface{}    `json:"toolUseResult"`
}

// ProcessedEntry represents a processed log entry for display
type ProcessedEntry struct {
	UUID         string
	ParentUUID   string
	IsSidechain  bool
	Type         string
	Timestamp    string
	RawTimestamp string // Keep the raw timestamp for comparisons
	Role         string
	Content      template.HTML
	ToolCalls    []ToolCall
	IsToolResult bool
	IsError      bool
	Children     []*ProcessedEntry
	Depth        int
	ToolResultID string // For matching tool results to tool calls
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID          string
	Name        string
	Description string
	Input       template.HTML
	Result      *ProcessedEntry // Tool result entry
	TaskEntries []*ProcessedEntry // For Task tool - sidechain entries
}

func main() {
	var inputFile, outputFile string
	var openBrowser bool
	flag.StringVar(&inputFile, "input", "", "Input JSONL file path")
	flag.StringVar(&outputFile, "output", "", "Output HTML file path (optional)")
	flag.BoolVar(&openBrowser, "open", false, "Open the generated HTML file in browser")
	flag.Parse()

	if inputFile == "" {
		log.Fatal("Please provide an input file using -input flag")
	}

	// If no output file specified, create a temp file and auto-open it
	autoOpen := false
	if outputFile == "" {
		// Generate unique filename based on input file and timestamp
		baseName := filepath.Base(inputFile)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		timestamp := time.Now().Format("20060102-150405")
		outputFile = filepath.Join(os.TempDir(), fmt.Sprintf("cclog-%s-%s.html", baseName, timestamp))
		autoOpen = true
	}

	entries, err := readJSONLFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	processed := processEntries(entries)
	
	err = generateHTML(processed, outputFile)
	if err != nil {
		log.Fatalf("Error generating HTML: %v", err)
	}

	fmt.Printf("Successfully generated %s\n", outputFile)
	
	// Open browser if -open flag was set OR if output was auto-generated
	if openBrowser || autoOpen {
		if err := openInBrowser(outputFile); err != nil {
			log.Printf("Warning: Could not open browser: %v", err)
		}
	}
}

func readJSONLFile(filename string) ([]LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max line size

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			log.Printf("Error parsing line %d: %v", lineNum, err)
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func processEntries(entries []LogEntry) []*ProcessedEntry {
	// Create a map for quick lookup
	entryMap := make(map[string]*ProcessedEntry)
	var rootEntries []*ProcessedEntry
	toolCallMap := make(map[string]*ToolCall) // Map tool ID to ToolCall

	// First pass: create all processed entries
	for _, entry := range entries {
		processed := processEntry(entry)
		entryMap[processed.UUID] = processed
		
		// Track tool calls for later matching
		for i := range processed.ToolCalls {
			toolCallMap[processed.ToolCalls[i].ID] = &processed.ToolCalls[i]
		}
	}

	// Second pass: build chronological list of main conversation entries
	for _, entry := range entries {
		if !entry.IsSidechain {
			processed := entryMap[entry.UUID]
			// If this is a tool result, attach it to the corresponding tool call
			if processed.IsToolResult && processed.ToolResultID != "" {
				if toolCall, exists := toolCallMap[processed.ToolResultID]; exists {
					toolCall.Result = processed
					continue // Don't add as a regular entry
				}
			}
			rootEntries = append(rootEntries, processed)
		}
	}

	// Third pass: group sidechain entries with their corresponding Task tool calls
	// Match Task tool calls with their sidechain entries based on timing
	var sidechainRoots []*ProcessedEntry
	for _, processed := range entryMap {
		if processed.IsSidechain && processed.ParentUUID == "" {
			sidechainRoots = append(sidechainRoots, processed)
		}
	}
	
	// Debug: log sidechain count
	if len(sidechainRoots) > 0 {
		log.Printf("Found %d sidechain root entries", len(sidechainRoots))
	}
	
	for _, sidechain := range sidechainRoots {
		// Find the most recent Task tool call before this sidechain entry
		var bestMatch *ToolCall
		var bestTimeStr string
		
		for _, entry := range entryMap {
			for i := range entry.ToolCalls {
				if entry.ToolCalls[i].Name == "Task" {
					// Compare raw timestamps
					if entry.RawTimestamp < sidechain.RawTimestamp {
						if bestMatch == nil || entry.RawTimestamp > bestTimeStr {
							bestMatch = &entry.ToolCalls[i]
							bestTimeStr = entry.RawTimestamp
							log.Printf("Found potential Task match: tool at %s for sidechain at %s", entry.RawTimestamp, sidechain.RawTimestamp)
						}
					}
				}
			}
		}
		
		// If we found a matching Task tool call, attach the sidechain entries
		if bestMatch != nil && len(bestMatch.TaskEntries) == 0 {
			bestMatch.TaskEntries = collectSidechainEntries(sidechain, entryMap)
			log.Printf("Attached %d sidechain entries to Task tool call", len(bestMatch.TaskEntries))
		} else if bestMatch == nil {
			log.Printf("No matching Task tool call found for sidechain at %s", sidechain.RawTimestamp)
		}
	}

	// No need to calculate depths for chronological display

	return rootEntries
}

func collectSidechainEntries(root *ProcessedEntry, entryMap map[string]*ProcessedEntry) []*ProcessedEntry {
	var result []*ProcessedEntry
	
	// Build the sidechain tree structure
	var buildTree func(entry *ProcessedEntry, depth int)
	buildTree = func(entry *ProcessedEntry, depth int) {
		entry.Depth = depth
		result = append(result, entry)
		
		// Find and add children
		for _, e := range entryMap {
			if e.ParentUUID == entry.UUID && e.IsSidechain {
				entry.Children = append(entry.Children, e)
			}
		}
		
		// Recursively process children
		for _, child := range entry.Children {
			buildTree(child, depth+1)
		}
	}
	
	buildTree(root, 0)
	return result
}

func calculateDepth(entry *ProcessedEntry, depth int) {
	entry.Depth = depth
	for _, child := range entry.Children {
		calculateDepth(child, depth+1)
	}
}

func processEntry(entry LogEntry) *ProcessedEntry {
	processed := &ProcessedEntry{
		UUID:         entry.UUID,
		IsSidechain:  entry.IsSidechain,
		Type:         entry.Type,
		Timestamp:    formatTimestamp(entry.Timestamp),
		RawTimestamp: entry.Timestamp,
	}

	if entry.ParentUUID != nil {
		processed.ParentUUID = *entry.ParentUUID
	}

	// Process the message content
	var msg map[string]interface{}
	if err := json.Unmarshal(entry.Message, &msg); err == nil {
		processed.Role = getStringValue(msg, "role")
		
		// Handle different message types
		switch processed.Type {
		case "user":
			processed.Content = processUserMessage(msg)
			processed.IsToolResult = isToolResult(msg)
		case "assistant":
			processed.Content, processed.ToolCalls = processAssistantMessage(msg)
		}
		
		// Check if it's an error and extract tool result ID
		if processed.IsToolResult {
			if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
				if toolResult, ok := content[0].(map[string]interface{}); ok {
					processed.IsError = getBoolValue(toolResult, "is_error")
					processed.ToolResultID = getStringValue(toolResult, "tool_use_id")
				}
			}
		}
	}

	return processed
}

func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("15:04:05")
}

func processUserMessage(msg map[string]interface{}) template.HTML {
	content := getStringValue(msg, "content")
	
	// Check if it's a tool result
	if contentArray, ok := msg["content"].([]interface{}); ok && len(contentArray) > 0 {
		if toolResult, ok := contentArray[0].(map[string]interface{}); ok {
			if toolType := getStringValue(toolResult, "type"); toolType == "tool_result" {
				// Handle different content types
				var toolContent string
				if contentVal, ok := toolResult["content"].(string); ok {
					toolContent = contentVal
				} else if contentArray, ok := toolResult["content"].([]interface{}); ok && len(contentArray) > 0 {
					// Handle array content (like from Task tool)
					if textContent, ok := contentArray[0].(map[string]interface{}); ok {
						toolContent = getStringValue(textContent, "text")
					}
				}
				
				isError := getBoolValue(toolResult, "is_error")
				
				if isError {
					return template.HTML(fmt.Sprintf(`<div class="tool-result error">%s</div>`, html.EscapeString(toolContent)))
				}
				return template.HTML(fmt.Sprintf(`<div class="tool-result">%s</div>`, formatContent(toolContent)))
			}
		}
	}
	
	return template.HTML(formatContent(content))
}

func processAssistantMessage(msg map[string]interface{}) (template.HTML, []ToolCall) {
	var content strings.Builder
	var toolCalls []ToolCall

	if contentArray, ok := msg["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if contentItem, ok := item.(map[string]interface{}); ok {
				contentType := getStringValue(contentItem, "type")
				
				switch contentType {
				case "text":
					text := getStringValue(contentItem, "text")
					if text != "" {
						content.WriteString(formatContent(text))
					}
				case "tool_use":
					tool := processToolUse(contentItem)
					toolCalls = append(toolCalls, tool)
				}
			}
		}
	}

	return template.HTML(content.String()), toolCalls
}

func processToolUse(toolUse map[string]interface{}) ToolCall {
	tool := ToolCall{
		ID:   getStringValue(toolUse, "id"),
		Name: getStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.Description = getStringValue(input, "description")
		
		// Format the input as JSON
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		tool.Input = template.HTML(fmt.Sprintf(`<pre class="tool-input">%s</pre>`, html.EscapeString(string(inputJSON))))
	}

	return tool
}

func formatContent(content string) string {
	// Escape HTML
	content = html.EscapeString(content)
	
	// Convert newlines to <br>
	content = strings.ReplaceAll(content, "\n", "<br>")
	
	// Wrap code blocks
	content = strings.ReplaceAll(content, "```", "</code></pre>CODE_BLOCK_MARKER<pre><code>")
	content = strings.ReplaceAll(content, "CODE_BLOCK_MARKER", "```")
	
	// Remove any empty pre/code tags at start/end
	content = strings.TrimPrefix(content, "</code></pre>```")
	content = strings.TrimSuffix(content, "```<pre><code>")
	
	return content
}

func isToolResult(msg map[string]interface{}) bool {
	if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
		if toolResult, ok := content[0].(map[string]interface{}); ok {
			return getStringValue(toolResult, "type") == "tool_result"
		}
	}
	return false
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func generateHTML(entries []*ProcessedEntry, outputFile string) error {
	tmplText := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Claude Code Log Viewer</title>
    <style>
        * {
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
        }
        
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        
        .entry {
            margin-bottom: 20px;
            border-left: 3px solid transparent;
            padding-left: 15px;
            transition: all 0.2s ease;
        }
        
        .entry.user {
            border-left-color: #3498db;
            background: #f0f7ff;
            padding: 10px 15px;
            border-radius: 4px;
        }
        
        .entry.assistant {
            border-left-color: #27ae60;
            background: #f0fff4;
            padding: 10px 15px;
            border-radius: 4px;
        }
        
        .entry.sidechain {
            opacity: 0.9;
        }
        
        .entry.sidechain.user {
            background: #f0fff4; /* Same green background as main assistant */
            border-left-color: #27ae60; /* Green border like main assistant */
        }
        
        .entry.sidechain.assistant {
            background: #fff3cd; /* Yellow background for sub-agent */
            border-left-color: #f39c12; /* Orange border for sub-agent */
        }
        
        .task-entry .entry.sidechain {
            margin-left: 0; /* Override any margin for entries within task */
        }
        
        .entry-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
            font-size: 0.9em;
            color: #666;
        }
        
        .role {
            font-weight: bold;
            text-transform: capitalize;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.85em;
        }
        
        .role.user {
            background: #3498db;
            color: white;
        }
        
        .role.assistant {
            background: #27ae60;
            color: white;
        }
        
        .role.subagent {
            background: #f39c12;
            color: white;
        }
        
        .timestamp {
            color: #999;
            font-size: 0.85em;
        }
        
        .content {
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        
        .tool-calls {
            margin-top: 10px;
        }
        
        .tool-call {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 10px;
            margin-bottom: 10px;
        }
        
        .tool-header {
            display: flex;
            align-items: center;
            gap: 10px;
            cursor: pointer;
            user-select: none;
        }
        
        .tool-header:hover {
            background: #e9ecef;
            margin: -10px;
            padding: 10px;
            border-radius: 4px;
        }
        
        .tool-name {
            font-weight: bold;
            color: #495057;
        }
        
        .tool-description {
            color: #6c757d;
            font-size: 0.9em;
        }
        
        .expand-icon {
            width: 20px;
            height: 20px;
            transition: transform 0.2s;
        }
        
        .expanded .expand-icon {
            transform: rotate(90deg);
        }
        
        .tool-details {
            display: none;
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px solid #dee2e6;
        }
        
        .expanded .tool-details {
            display: block;
        }
        
        .tool-input {
            background: #f1f3f5;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
            font-size: 0.85em;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .tool-result {
            background: #e3f2fd;
            padding: 10px;
            border-radius: 4px;
            margin-top: 5px;
            white-space: pre-wrap;
        }
        
        .tool-result.error {
            background: #ffebee;
            color: #c62828;
            border: 1px solid #ef5350;
        }
        
        .children {
            margin-left: 20px;
            margin-top: 15px;
            padding-left: 20px;
            border-left: 2px dashed #ddd;
        }
        
        .task-children {
            background: #fafafa;
            border-radius: 4px;
            padding: 10px;
            margin-top: 10px;
        }
        
        pre {
            margin: 0;
        }
        
        code {
            background: #f4f4f4;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        pre code {
            background: none;
            padding: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Claude Code Conversation Log</h1>
        {{range .}}
            {{template "entry" .}}
        {{end}}
    </div>
    
    <script>
        // Toggle tool call details
        document.querySelectorAll('.tool-header').forEach(header => {
            header.addEventListener('click', () => {
                header.parentElement.classList.toggle('expanded');
            });
        });
        
        // Toggle result sections
        document.querySelectorAll('.result-header').forEach(header => {
            header.addEventListener('click', () => {
                const icon = header.querySelector('.result-expand-icon');
                const content = header.nextElementSibling;
                if (content) {
                    const isHidden = content.style.display === 'none';
                    content.style.display = isHidden ? 'block' : 'none';
                    icon.style.transform = isHidden ? 'rotate(90deg)' : 'rotate(0deg)';
                }
            });
        });
        
    </script>
</body>
</html>

{{define "entry"}}
<div class="entry {{.Type}}{{if .IsSidechain}} sidechain{{end}}">
    <div class="entry-header">
        {{if .IsSidechain}}
            {{if eq .Role "user"}}
            <span class="role assistant">Assistant</span>
            {{else if eq .Role "assistant"}}
            <span class="role subagent">Sub Agent</span>
            {{end}}
        {{else}}
            <span class="role {{.Role}}">{{.Role}}</span>
        {{end}}
        <span class="timestamp">{{.Timestamp}}</span>
        {{if .IsSidechain}}
        <span style="color: #9c27b0; font-size: 0.85em;">📎 Task</span>
        {{end}}
    </div>
    
    <div class="content">{{.Content}}</div>
    
    {{if .ToolCalls}}
    <div class="tool-calls">
        {{range .ToolCalls}}
        <div class="tool-call">
            <div class="tool-header">
                <svg class="expand-icon" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="tool-name">{{.Name}}</span>
                {{if .Description}}
                <span class="tool-description">{{.Description}}</span>
                {{end}}
            </div>
            <div class="tool-details">
                {{.Input}}
                {{if .TaskEntries}}
                <div style="margin-top: 15px;">
                    {{range .TaskEntries}}
                        <div class="task-entry">
                            {{template "entry" .}}
                        </div>
                    {{end}}
                </div>
                {{end}}
                {{if .Result}}
                <div class="tool-result-section" style="margin-top: 15px;">
                    <div class="result-header" style="cursor: pointer; user-select: none; display: flex; align-items: center; gap: 5px;">
                        <svg class="result-expand-icon" width="16" height="16" viewBox="0 0 20 20" fill="currentColor" style="transition: transform 0.2s;">
                            <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                        </svg>
                        <strong>Result</strong>
                    </div>
                    <div class="result-content" style="display: none; margin-top: 10px;">
                        {{.Result.Content}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
    </div>
    {{end}}
    
</div>
{{end}}`

	// Create custom function map
	funcMap := template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, entries)
}

func openInBrowser(filename string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filename)
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filename)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	
	return cmd.Start()
}