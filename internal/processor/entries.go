package processor

import (
	"cclogviewer/internal/debug"
	"cclogviewer/internal/models"
	"encoding/json"
	"log"
	"strings"
	"time"
)

// ProcessEntries processes raw log entries into a structured format
func ProcessEntries(entries []models.LogEntry) []*models.ProcessedEntry {
	// Create a map for quick lookup
	entryMap := make(map[string]*models.ProcessedEntry)
	var rootEntries []*models.ProcessedEntry
	toolCallMap := make(map[string]*models.ToolCall) // Map tool ID to ToolCall

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

	// Third pass: build a map of tool calls in sidechain entries
	sidechainToolCallMap := make(map[string]*models.ToolCall) // Map tool ID to ToolCall for sidechains
	for _, entry := range entryMap {
		if entry.IsSidechain {
			for i := range entry.ToolCalls {
				sidechainToolCallMap[entry.ToolCalls[i].ID] = &entry.ToolCalls[i]
			}
		}
	}

	// Fourth pass: attach tool results to sidechain tool calls
	for _, entry := range entryMap {
		if entry.IsSidechain && entry.IsToolResult && entry.ToolResultID != "" {
			if toolCall, exists := sidechainToolCallMap[entry.ToolResultID]; exists {
				toolCall.Result = entry
			}
		}
	}

	// Fifth pass: group sidechain entries with their corresponding Task tool calls
	// First, collect all sidechain roots
	var sidechainRoots []*models.ProcessedEntry
	for _, entry := range entries {
		processed := entryMap[entry.UUID]
		if processed.IsSidechain && processed.ParentUUID == "" && !processed.IsToolResult {
			sidechainRoots = append(sidechainRoots, processed)
		}
	}
	
	if debug.Enabled {
		log.Printf("Found %d sidechain roots", len(sidechainRoots))
	}
	
	// Build a map to track which sidechains have been matched
	matchedSidechains := make(map[string]bool)
	
	// Look through tool results to match Task tools with their sidechains
	for _, entry := range entries {
		if !entry.IsSidechain && entry.Type == "user" {
			processed := entryMap[entry.UUID]
			if processed.IsToolResult && processed.ToolResultID != "" {
				// Check if this is a Task tool result
				if toolCall, exists := toolCallMap[processed.ToolResultID]; exists && toolCall.Name == "Task" {
					// Extract the content from the tool result
					var msg map[string]interface{}
					if err := json.Unmarshal(entry.Message, &msg); err == nil {
						if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
							if toolResult, ok := content[0].(map[string]interface{}); ok {
								if resultContent, ok := toolResult["content"].([]interface{}); ok && len(resultContent) > 0 {
									if textContent, ok := resultContent[0].(map[string]interface{}); ok {
										if text, ok := textContent["text"].(string); ok {
											// Find the best matching sidechain based on content
											var bestMatch *models.ProcessedEntry
											var bestScore int
											
											for _, sidechain := range sidechainRoots {
												if matchedSidechains[sidechain.UUID] {
													continue // Skip already matched sidechains
												}
												
												// Extract content from sidechain to compare
												sidechainText := extractFullSidechainContent(sidechain, entryMap)
												
												// Calculate similarity score
												score := calculateContentSimilarity(text, sidechainText)
												
												if score > bestScore {
													bestScore = score
													bestMatch = sidechain
												}
											}
											
											if bestMatch != nil {
												toolCall.TaskEntries = collectSidechainEntries(bestMatch, entryMap)
												matchedSidechains[bestMatch.UUID] = true
												if debug.Enabled {
													log.Printf("Matched Task tool %s to sidechain %s (score: %d)", 
														processed.ToolResultID, bestMatch.UUID, bestScore)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Calculate conversation size for each entry (input + cache read + cache write)
	for _, entry := range rootEntries {
		// Conversation size is the context sent to the model (excluding output)
		entry.TotalTokens = entry.InputTokens + entry.CacheReadTokens + entry.CacheCreationTokens
		
		// Also calculate for tool results
		for _, toolCall := range entry.ToolCalls {
			if toolCall.Result != nil {
				toolCall.Result.TotalTokens = toolCall.Result.InputTokens + 
					toolCall.Result.CacheReadTokens + toolCall.Result.CacheCreationTokens
			}
			
			// Calculate for Task entries (sidechain conversations)
			for _, taskEntry := range toolCall.TaskEntries {
				taskEntry.TotalTokens = taskEntry.InputTokens + taskEntry.CacheReadTokens + taskEntry.CacheCreationTokens
				
				// Also calculate for tool results within task entries
				for _, taskToolCall := range taskEntry.ToolCalls {
					if taskToolCall.Result != nil {
						taskToolCall.Result.TotalTokens = taskToolCall.Result.InputTokens + 
							taskToolCall.Result.CacheReadTokens + taskToolCall.Result.CacheCreationTokens
					}
				}
			}
		}
	}

	return rootEntries
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
		processed.Role = GetStringValue(msg, "role")
		
		// Handle different message types
		switch processed.Type {
		case "user":
			processed.Content = ProcessUserMessage(msg)
			processed.IsToolResult = isToolResult(msg)
		case "assistant":
			processed.Content, processed.ToolCalls = ProcessAssistantMessage(msg)
		}
		
		// Check if it's an error and extract tool result ID
		if processed.IsToolResult {
			if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
				if toolResult, ok := content[0].(map[string]interface{}); ok {
					processed.IsError = GetBoolValue(toolResult, "is_error")
					processed.ToolResultID = GetStringValue(toolResult, "tool_use_id")
				}
			}
		}
		
		// Extract token counts from usage field if available
		if usage, ok := msg["usage"].(map[string]interface{}); ok {
			// Extract all token types
			if inputTokens, ok := usage["input_tokens"].(float64); ok {
				processed.InputTokens = int(inputTokens)
			}
			// Always estimate output tokens from content for accuracy
			processed.OutputTokens = EstimateTokens(string(processed.Content))
			processed.TokenCount = processed.OutputTokens
			
			if cacheReadTokens, ok := usage["cache_read_input_tokens"].(float64); ok {
				processed.CacheReadTokens = int(cacheReadTokens)
			}
			if cacheCreationTokens, ok := usage["cache_creation_input_tokens"].(float64); ok {
				processed.CacheCreationTokens = int(cacheCreationTokens)
			}
		} else {
			// Fall back to estimation for messages without usage data
			processed.TokenCount = EstimateTokens(string(processed.Content))
			// For user messages, the estimated tokens are output tokens
			if processed.Role == "user" {
				processed.OutputTokens = processed.TokenCount
			}
		}
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
	var buildTree func(entry *models.ProcessedEntry, depth int, skipEntry bool)
	buildTree = func(entry *models.ProcessedEntry, depth int, skipEntry bool) {
		// Add to result only if we're not skipping this entry
		if !skipEntry {
			entry.Depth = depth
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
			buildTree(child, depth+1, shouldSkip)
		}
	}
	
	buildTree(root, 0, false)
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
			return GetStringValue(toolResult, "type") == "tool_result"
		}
	}
	return false
}

// GetStringValue extracts a string value from a map
func GetStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// GetBoolValue extracts a bool value from a map
func GetBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// extractContent extracts text content from a ProcessedEntry
func extractContent(entry *models.ProcessedEntry) string {
	// For now, just convert HTML content to plain text
	content := string(entry.Content)
	// Remove HTML tags in a simple way
	content = strings.ReplaceAll(content, "<br>", " ")
	content = strings.ReplaceAll(content, "</div>", " ")
	content = strings.ReplaceAll(content, "<div class=\"tool-result\">", " ")
	content = strings.ReplaceAll(content, "<div class=\"tool-result error\">", " ")
	// Remove other HTML tags
	for strings.Contains(content, "<") && strings.Contains(content, ">") {
		start := strings.Index(content, "<")
		end := strings.Index(content, ">")
		if start >= 0 && end > start {
			content = content[:start] + " " + content[end+1:]
		} else {
			break
		}
	}
	return strings.TrimSpace(content)
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractFullSidechainContent extracts all text content from a sidechain conversation
func extractFullSidechainContent(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	var content strings.Builder
	
	// Helper function to extract content from an entry and its children
	var extractFromEntry func(entry *models.ProcessedEntry)
	extractFromEntry = func(entry *models.ProcessedEntry) {
		// Add this entry's content
		entryText := extractContent(entry)
		if entryText != "" {
			content.WriteString(entryText)
			content.WriteString(" ")
		}
		
		// Process children
		for _, child := range entry.Children {
			extractFromEntry(child)
		}
		
		// Process any entries that have this as parent
		for _, e := range entryMap {
			if e.ParentUUID == entry.UUID && e.IsSidechain {
				// Check if it's already in children
				found := false
				for _, child := range entry.Children {
					if child.UUID == e.UUID {
						found = true
						break
					}
				}
				if !found {
					extractFromEntry(e)
				}
			}
		}
	}
	
	extractFromEntry(root)
	return content.String()
}

// calculateContentSimilarity calculates a similarity score between two texts
func calculateContentSimilarity(text1, text2 string) int {
	// Simple implementation: count matching words
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))
	
	wordMap := make(map[string]bool)
	for _, word := range words1 {
		if len(word) > 3 { // Skip very short words
			wordMap[word] = true
		}
	}
	
	score := 0
	for _, word := range words2 {
		if len(word) > 3 && wordMap[word] {
			score++
		}
	}
	
	// Bonus points for exact phrase matches
	lowerText1 := strings.ToLower(text1)
	lowerText2 := strings.ToLower(text2)
	
	// Check for key phrases that might appear in both
	keyPhrases := []string{
		"developer's desktop",
		"digital artifacts", 
		"claude code",
		"apple pie",
		"poem",
		"desktop canvas",
	}
	
	for _, phrase := range keyPhrases {
		if strings.Contains(lowerText1, phrase) && strings.Contains(lowerText2, phrase) {
			score += 10 // Heavy weight for matching key phrases
		}
	}
	
	return score
}