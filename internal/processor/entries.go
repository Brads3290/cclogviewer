package processor

import (
	"encoding/json"
	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/utils"
	"strings"
	"time"
)

// ProcessEntries builds a hierarchical structure from flat log entries.
func ProcessEntries(entries []models.LogEntry) []*models.ProcessedEntry {
	state := initializeProcessingState(len(entries))
	entryMap := make(map[string]*models.ProcessedEntry)

	// Phase 1: Process all entries
	processAllEntries(entries, state, entryMap)

	// Phase 2: Match tool calls with results
	matchToolCallsWithResults(state.Entries)

	// Phase 3: Process sidechains
	processSidechainConversations(state, entries, entryMap)

	// Phase 4-7: Post-processing
	rootEntries := getRootEntries(state)
	calculateAllTokens(rootEntries)
	checkAllMissingResults(rootEntries)
	linkAllCommandOutputs(rootEntries)
	buildFinalHierarchy(rootEntries)

	return rootEntries
}

// checkMissingToolResults recursively checks for missing tool results and sidechains
func checkMissingToolResults(entry *models.ProcessedEntry) {
	// Check tool calls in this entry
	for i := range entry.ToolCalls {
		toolCall := &entry.ToolCalls[i]

		// Check if result is missing
		if toolCall.Result == nil {
			toolCall.HasMissingResult = true
		}

		// For Task tools, also check if sidechain is missing
		if toolCall.Name == ToolNameTask && len(toolCall.TaskEntries) == 0 {
			toolCall.HasMissingSidechain = true
		}

		// Recursively check Task entries
		for _, taskEntry := range toolCall.TaskEntries {
			checkMissingToolResults(taskEntry)
		}
	}

	// Recursively check children
	for _, child := range entry.Children {
		checkMissingToolResults(child)
	}
}

// calculateTokensForEntry recursively aggregates tokens across nested tool calls.
func calculateTokensForEntry(entry *models.ProcessedEntry) {
	entry.TotalTokens = entry.InputTokens + entry.OutputTokens +
		entry.CacheReadTokens + entry.CacheCreationTokens

	// Calculate for tool calls
	for i := range entry.ToolCalls {
		toolCall := &entry.ToolCalls[i]

		// Calculate for tool result
		if toolCall.Result != nil {
			toolCall.Result.TotalTokens = toolCall.Result.InputTokens +
				toolCall.Result.OutputTokens +
				toolCall.Result.CacheReadTokens + toolCall.Result.CacheCreationTokens
		}

		// Recursively calculate for nested Task entries
		for _, taskEntry := range toolCall.TaskEntries {
			calculateTokensForEntry(taskEntry)
		}
	}
}

func processEntry(entry models.LogEntry) *models.ProcessedEntry {
	processed := &models.ProcessedEntry{
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
		// Process message using handlers
		if err := processMessage(processed, msg, entry); err != nil {
			// Log error but continue processing
		}

		// Process token counts
		tokenProcessor := NewTokenProcessor()
		tokenProcessor.ProcessTokens(processed, msg)
	} else {
		// If we can't parse the message, estimate tokens from content
		processed.TokenCount = EstimateTokens(string(processed.Content))
	}

	return processed
}

func collectSidechainEntries(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) []*models.ProcessedEntry {
	var result []*models.ProcessedEntry

	// First, collect all tool results that are attached to tool calls
	attachedToolResults := make(map[string]bool)
	for _, entry := range entryMap {
		if entry.IsSidechain {
			for _, toolCall := range entry.ToolCalls {
				if toolCall.Result != nil {
					attachedToolResults[toolCall.Result.UUID] = true
				}
			}
		}
	}

	// Build the sidechain tree structure
	var buildTree func(entry *models.ProcessedEntry, skipEntry bool)
	buildTree = func(entry *models.ProcessedEntry, skipEntry bool) {
		// Add to result only if we're not skipping this entry
		if !skipEntry {
			result = append(result, entry)
		}

		// Find and add children
		for _, e := range entryMap {
			if e.ParentUUID == entry.UUID && e.IsSidechain {
				entry.Children = append(entry.Children, e)
			}
		}

		// Recursively process children
		for _, child := range entry.Children {
			// Skip tool results that have been attached to tool calls when adding to result,
			// but still process their children
			shouldSkip := child.IsToolResult && attachedToolResults[child.UUID]
			buildTree(child, shouldSkip)
		}
	}

	buildTree(root, false)
	return result
}

func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("15:04:05")
}

func isToolResult(msg map[string]interface{}) bool {
	if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
		if toolResult, ok := content[0].(map[string]interface{}); ok {
			return utils.ExtractString(toolResult, "type") == TypeToolResult
		}
	}
	return false
}



// extractContent extracts text content from a ProcessedEntry
func extractContent(entry *models.ProcessedEntry) string {
	// Content is now stored as raw text, no HTML processing needed
	return strings.TrimSpace(entry.Content)
}


// getFirstUserMessage finds the first user message in a sidechain conversation
func getFirstUserMessage(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	// First check if root itself is a user message
	if root.Role == RoleUser {
		return extractContent(root)
	}

	// Otherwise, look for the first user message in the tree
	var findFirstUser func(entry *models.ProcessedEntry) string
	findFirstUser = func(entry *models.ProcessedEntry) string {
		// Check children first (in order)
		for _, child := range entry.Children {
			if child.Role == RoleUser {
				return extractContent(child)
			}
		}

		// Then recursively check children's children
		for _, child := range entry.Children {
			if result := findFirstUser(child); result != "" {
				return result
			}
		}

		// Also check entries that have this as parent
		for _, e := range entryMap {
			if e.ParentUUID == entry.UUID && e.IsSidechain && e.Role == RoleUser {
				return extractContent(e)
			}
		}

		return ""
	}

	return findFirstUser(root)
}

// getLastAssistantMessage finds the last assistant message in a sidechain conversation
func getLastAssistantMessage(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	var lastAssistantContent string
	var lastAssistantTime time.Time

	var findLastAssistant func(entry *models.ProcessedEntry)
	findLastAssistant = func(entry *models.ProcessedEntry) {
		// Check if this is an assistant message
		if entry.Role == RoleAssistant && !entry.IsToolResult {
			// Parse timestamp
			if t, err := time.Parse(time.RFC3339, entry.RawTimestamp); err == nil {
				if lastAssistantContent == "" || t.After(lastAssistantTime) {
					lastAssistantContent = extractContent(entry)
					lastAssistantTime = t
				}
			}
		}

		// Check children
		for _, child := range entry.Children {
			findLastAssistant(child)
		}

		// Check entries that have this as parent
		for _, e := range entryMap {
			if e.ParentUUID == entry.UUID && e.IsSidechain {
				// Avoid infinite recursion by checking if already in children
				found := false
				for _, child := range entry.Children {
					if child.UUID == e.UUID {
						found = true
						break
					}
				}
				if !found {
					findLastAssistant(e)
				}
			}
		}
	}

	findLastAssistant(root)
	return lastAssistantContent
}

// normalizeText normalizes text for comparison by removing extra whitespace and newlines
func normalizeText(text string) string {
	// Replace all newlines with spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

// extractXMLContent extracts content between XML tags
func extractXMLContent(text, tag string) string {
	startTag := "<" + tag + ">"
	endTag := "</" + tag + ">"

	startIdx := strings.Index(text, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(text[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return text[startIdx : startIdx+endIdx]
}

// isCommandWithStdout checks if current entry is a command message and next entry contains stdout
func isCommandWithStdout(current, next *models.ProcessedEntry) bool {
	return current.IsCommandMessage &&
		next.Role == RoleUser &&
		strings.Contains(next.Content, "<"+TagCommandStdout+">")
}

// linkCommandOutputs merges adjacent command and stdout messages for cleaner display.
func linkCommandOutputs(entries []*models.ProcessedEntry) {
	for i := 0; i < len(entries)-1; i++ {
		current := entries[i]
		next := entries[i+1]

		// If current is a command message and next contains stdout
		if isCommandWithStdout(current, next) {
			// Extract the stdout content
			current.CommandOutput = extractXMLContent(next.Content, TagCommandStdout)
			// Mark the next entry for removal
			next.Content = ""
		}
	}
}
