package processor

import (
	"cclogviewer/internal/models"
	"encoding/json"
	"log"
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
	// Match Task tool calls with their sidechain entries based on timing
	var sidechainRoots []*models.ProcessedEntry
	for _, processed := range entryMap {
		if processed.IsSidechain && processed.ParentUUID == "" {
			// Skip tool results that are attached to tool calls
			if processed.IsToolResult {
				continue
			}
			sidechainRoots = append(sidechainRoots, processed)
		}
	}
	
	// Debug: log sidechain count
	if len(sidechainRoots) > 0 {
		log.Printf("Found %d sidechain root entries", len(sidechainRoots))
	}
	
	for _, sidechain := range sidechainRoots {
		// Find the most recent Task tool call before this sidechain entry
		var bestMatch *models.ToolCall
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
			for _, entry := range bestMatch.TaskEntries {
				log.Printf("  - Entry: UUID=%s, Role=%s, IsToolResult=%v", entry.UUID, entry.Role, entry.IsToolResult)
			}
		} else if bestMatch == nil {
			log.Printf("No matching Task tool call found for sidechain at %s", sidechain.RawTimestamp)
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