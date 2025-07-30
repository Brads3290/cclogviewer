package processor

import (
	"encoding/json"
	"github.com/Brads3290/cclogviewer/internal/debug"
	"github.com/Brads3290/cclogviewer/internal/models"
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
					// Check if the tool was interrupted
					if processed.IsError && strings.Contains(strings.ToLower(processed.Content), "request interrupted by user") {
						toolCall.IsInterrupted = true
					}
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
				// Check if the tool was interrupted
				if entry.IsError && strings.Contains(strings.ToLower(entry.Content), "request interrupted by user") {
					toolCall.IsInterrupted = true
				}
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

	// Look through tool calls to match Task tools with their sidechains
	// Process both main conversation and sidechain entries to handle nested Tasks
	for _, entry := range entries {
		if entry.Type == "assistant" {
			processed := entryMap[entry.UUID]
			
			if debug.Enabled && len(processed.ToolCalls) > 0 {
				log.Printf("Processing assistant entry %s (sidechain: %v) with %d tool calls", 
					entry.UUID, entry.IsSidechain, len(processed.ToolCalls))
			}
			
			for i := range processed.ToolCalls {
				toolCall := &processed.ToolCalls[i]
				if toolCall.Name == "Task" {
					if debug.Enabled {
						log.Printf("Found Task tool %s in entry %s (sidechain: %v)", 
							toolCall.ID, entry.UUID, entry.IsSidechain)
					}
					
					// Extract the prompt from the Task tool call
					taskPrompt := extractTaskPrompt(toolCall)
					if taskPrompt == "" {
						if debug.Enabled {
							log.Printf("Task tool %s has empty prompt, skipping", toolCall.ID)
						}
						continue
					}
					
					if debug.Enabled {
						log.Printf("Task tool %s prompt: %.50s...", toolCall.ID, taskPrompt)
					}

					// Extract the result text from the tool result
					var taskResult string
					if toolCall.Result != nil {
						// Find the original entry for this result
						for _, e := range entries {
							if e.UUID == toolCall.Result.UUID {
								var msg map[string]interface{}
								if err := json.Unmarshal(e.Message, &msg); err == nil {
									if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
										if toolResult, ok := content[0].(map[string]interface{}); ok {
											if resultContent, ok := toolResult["content"].([]interface{}); ok && len(resultContent) > 0 {
												if textContent, ok := resultContent[0].(map[string]interface{}); ok {
													if text, ok := textContent["text"].(string); ok {
														taskResult = text
													}
												}
											}
										}
									}
								}
								break
							}
						}
					}

					if taskResult == "" {
						if debug.Enabled {
							log.Printf("Task tool %s has empty result, skipping", toolCall.ID)
						}
						continue
					}
					
					if debug.Enabled {
						log.Printf("Task tool %s result: %.50s...", toolCall.ID, taskResult)
					}

					// Find the best matching sidechain
					var bestMatch *models.ProcessedEntry
					var bestMatchScore int

					if debug.Enabled {
						log.Printf("Searching among %d sidechain roots for Task %s", 
							len(sidechainRoots), toolCall.ID)
					}

					for _, sidechain := range sidechainRoots {
						if matchedSidechains[sidechain.UUID] {
							continue // Skip already matched sidechains
						}

						// Get first user message and last assistant message from sidechain
						firstUser := getFirstUserMessage(sidechain, entryMap)
						lastAssistant := getLastAssistantMessage(sidechain, entryMap)

						if firstUser == "" || lastAssistant == "" {
							if debug.Enabled {
								log.Printf("Sidechain %s has empty first user or last assistant, skipping", 
									sidechain.UUID)
							}
							continue
						}
						
						if debug.Enabled {
							log.Printf("Comparing Task %s with sidechain %s:", toolCall.ID, sidechain.UUID)
							log.Printf("  Task prompt: %.50s...", taskPrompt)
							log.Printf("  First user:  %.50s...", firstUser)
							log.Printf("  Task result: %.50s...", taskResult)
							log.Printf("  Last asst:   %.50s...", lastAssistant)
							
							// Log length info
							log.Printf("  Prompt lengths: task=%d, user=%d", len(taskPrompt), len(firstUser))
							log.Printf("  Result lengths: task=%d, asst=%d", len(taskResult), len(lastAssistant))
							
							// Log whether they start with the same text
							promptStarts := strings.HasPrefix(strings.TrimSpace(firstUser), strings.TrimSpace(taskPrompt))
							resultStarts := strings.HasPrefix(strings.TrimSpace(lastAssistant), strings.TrimSpace(taskResult))
							log.Printf("  Prompt starts with match: %v, Result starts with match: %v", promptStarts, resultStarts)
							
							// Log first 100 chars of normalized text for debugging
							if toolCall.ID == "toolu_01TRNMekrBthfccxqpoeViRw" && sidechain.UUID == "c871d1a9-f5a8-4236-83f5-05f08b75cb38" {
								taskPromptNorm := normalizeText(taskPrompt)
								firstUserNorm := normalizeText(firstUser)
								taskResultNorm := normalizeText(taskResult)
								lastAssistantNorm := normalizeText(lastAssistant)
								log.Printf("  NORMALIZED Task prompt (first 100): %.100s", taskPromptNorm)
								log.Printf("  NORMALIZED First user  (first 100): %.100s", firstUserNorm)
								log.Printf("  Normalized lengths: task=%d, user=%d", len(taskPromptNorm), len(firstUserNorm))
								log.Printf("  Prefix match check: %v", strings.HasPrefix(firstUserNorm, taskPromptNorm) || strings.HasPrefix(taskPromptNorm, firstUserNorm))
								log.Printf("  Result prefix match check: %v", strings.HasPrefix(lastAssistantNorm, taskResultNorm) || strings.HasPrefix(taskResultNorm, lastAssistantNorm))
							}
						}

						// Calculate match score based on both prompt and result matching
						// Normalize texts for comparison (remove extra whitespace, newlines)
						taskPromptNorm := normalizeText(taskPrompt)
						firstUserNorm := normalizeText(firstUser)
						taskResultNorm := normalizeText(taskResult)
						lastAssistantNorm := normalizeText(lastAssistant)
						
						// Check for exact match first
						promptMatch := taskPromptNorm == firstUserNorm
						resultMatch := taskResultNorm == lastAssistantNorm
						
						// If not exact match, check if one starts with the other (for truncated content)
						if !promptMatch && len(taskPromptNorm) > 20 && len(firstUserNorm) > 20 {
							// Check if either is a prefix of the other
							promptMatch = strings.HasPrefix(firstUserNorm, taskPromptNorm) || 
								         strings.HasPrefix(taskPromptNorm, firstUserNorm)
						}
						if !resultMatch && len(taskResultNorm) > 20 && len(lastAssistantNorm) > 20 {
							// Check if either is a prefix of the other
							resultMatch = strings.HasPrefix(lastAssistantNorm, taskResultNorm) || 
								         strings.HasPrefix(taskResultNorm, lastAssistantNorm)
						}

						// Score: 2 points for both matching, 1 point for partial match
						score := 0
						if promptMatch {
							score++
						}
						if resultMatch {
							score++
						}
						
						if debug.Enabled {
							log.Printf("  Prompt match: %v, Result match: %v, Score: %d", 
								promptMatch, resultMatch, score)
						}

						if score > bestMatchScore {
							bestMatchScore = score
							bestMatch = sidechain
						}

						// If we have a perfect match (both prompt and result), we can stop looking
						if score == 2 {
							break
						}
					}

					if bestMatch != nil {
						toolCall.TaskEntries = collectSidechainEntries(bestMatch, entryMap)
						matchedSidechains[bestMatch.UUID] = true
						if debug.Enabled {
							log.Printf("Matched Task tool %s to sidechain %s (score: %d, entries: %d)",
								toolCall.ID, bestMatch.UUID, bestMatchScore, len(toolCall.TaskEntries))
						}
					} else {
						if debug.Enabled {
							log.Printf("No match found for Task tool %s", toolCall.ID)
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
				calculateTokensForEntry(taskEntry)
			}
		}
	}

	// Final pass: check for missing results and sidechains
	for _, entry := range rootEntries {
		checkMissingToolResults(entry)
	}

	// Link command stdout messages to their command messages
	linkCommandOutputs(rootEntries)

	// Set depth for all entries based on sidechain hierarchy
	// Root conversation starts at depth 1
	for _, entry := range rootEntries {
		setEntryDepth(entry, 1)
	}

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
		if toolCall.Name == "Task" && len(toolCall.TaskEntries) == 0 {
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

// calculateTokensForEntry recursively calculates tokens for an entry and all its nested tool calls
func calculateTokensForEntry(entry *models.ProcessedEntry) {
	entry.TotalTokens = entry.InputTokens + entry.CacheReadTokens + entry.CacheCreationTokens

	// Calculate for tool calls
	for i := range entry.ToolCalls {
		toolCall := &entry.ToolCalls[i]

		// Calculate for tool result
		if toolCall.Result != nil {
			toolCall.Result.TotalTokens = toolCall.Result.InputTokens +
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
		processed.Role = GetStringValue(msg, "role")

		// Handle different message types
		switch processed.Type {
		case "user":
			processed.Content = ProcessUserMessage(msg)
			processed.IsToolResult = isToolResult(msg)
			
			// Check if this is a caveat message from local commands
			if strings.HasPrefix(processed.Content, "Caveat: The messages below were generated by the user while running local commands.") {
				processed.IsCaveatMessage = true
			}
			
			// Check if this is a command message with XML syntax
			if strings.Contains(processed.Content, "<command-name>") && strings.Contains(processed.Content, "</command-name>") {
				processed.IsCommandMessage = true
				// Parse command details
				processed.CommandName = extractXMLContent(processed.Content, "command-name")
				processed.CommandArgs = extractXMLContent(processed.Content, "command-args")
			}
		case "assistant":
			processed.Content, processed.ToolCalls = ProcessAssistantMessage(msg, entry.CWD)
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
	// Content is now stored as raw text, no HTML processing needed
	return strings.TrimSpace(entry.Content)
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

// getFirstUserMessage finds the first user message in a sidechain conversation
func getFirstUserMessage(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	// First check if root itself is a user message
	if root.Role == "user" {
		return extractContent(root)
	}

	// Otherwise, look for the first user message in the tree
	var findFirstUser func(entry *models.ProcessedEntry) string
	findFirstUser = func(entry *models.ProcessedEntry) string {
		// Check children first (in order)
		for _, child := range entry.Children {
			if child.Role == "user" {
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
			if e.ParentUUID == entry.UUID && e.IsSidechain && e.Role == "user" {
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
		if entry.Role == "assistant" && !entry.IsToolResult {
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

// extractTaskPrompt extracts the prompt from a Task tool call's raw input
func extractTaskPrompt(toolCall *models.ToolCall) string {
	if toolCall.RawInput == nil {
		return ""
	}

	inputMap, ok := toolCall.RawInput.(map[string]interface{})
	if !ok {
		return ""
	}

	prompt, ok := inputMap["prompt"].(string)
	if !ok {
		return ""
	}

	return prompt
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

// linkCommandOutputs links stdout messages to their preceding command messages
func linkCommandOutputs(entries []*models.ProcessedEntry) {
	for i := 0; i < len(entries)-1; i++ {
		current := entries[i]
		next := entries[i+1]
		
		// If current is a command message and next contains stdout
		if current.IsCommandMessage && next.Role == "user" && 
		   strings.Contains(next.Content, "<local-command-stdout>") {
			// Extract the stdout content
			current.CommandOutput = extractXMLContent(next.Content, "local-command-stdout")
			// Mark the next entry for removal
			next.Content = ""
		}
	}
}

// setEntryDepth recursively sets the depth for entries based on sidechain hierarchy
func setEntryDepth(entry *models.ProcessedEntry, depth int) {
	// Set the depth for this entry
	entry.Depth = depth
	
	// Process all tool calls
	for i := range entry.ToolCalls {
		toolCall := &entry.ToolCalls[i]
		
		// If this is a Task tool with sidechain entries, set their depth to current depth + 1
		if toolCall.Name == "Task" && len(toolCall.TaskEntries) > 0 {
			for _, taskEntry := range toolCall.TaskEntries {
				setEntryDepth(taskEntry, depth+1)
			}
		}
		
		// Also set depth for tool results
		if toolCall.Result != nil {
			toolCall.Result.Depth = depth
		}
	}
	
	// Process children (though main conversation entries shouldn't have children)
	for _, child := range entry.Children {
		setEntryDepth(child, depth)
	}
}

