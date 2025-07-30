package processor

import (
	"encoding/json"
	"github.com/Brads3290/cclogviewer/internal/debug"
	"github.com/Brads3290/cclogviewer/internal/models"
	"log"
	"strings"
)

// SidechainProcessor handles processing of sidechain conversations
type SidechainProcessor struct{}

// NewSidechainProcessor creates a new sidechain processor
func NewSidechainProcessor() *SidechainProcessor {
	return &SidechainProcessor{}
}

// ProcessSidechains processes sidechain conversations and matches them with Task tool calls
func (s *SidechainProcessor) ProcessSidechains(entries []*models.ProcessedEntry, originalEntries []models.LogEntry, entryMap map[string]*models.ProcessedEntry) error {
	// First, collect all sidechain roots
	sidechainRoots := s.collectSidechainRoots(originalEntries, entryMap)

	if debug.Enabled {
		log.Printf("Found %d sidechain roots", len(sidechainRoots))
	}

	// Build a map to track which sidechains have been matched
	matchedSidechains := make(map[string]bool)

	// Look through tool calls to match Task tools with their sidechains
	for _, entry := range originalEntries {
		if entry.Type == "assistant" {
			processed := entryMap[entry.UUID]

			if debug.Enabled && len(processed.ToolCalls) > 0 {
				log.Printf("Processing assistant entry %s (sidechain: %v) with %d tool calls",
					entry.UUID, entry.IsSidechain, len(processed.ToolCalls))
			}

			for i := range processed.ToolCalls {
				toolCall := &processed.ToolCalls[i]
				if toolCall.Name == "Task" {
					s.matchTaskWithSidechain(toolCall, &entry, originalEntries, sidechainRoots, entryMap, matchedSidechains)
				}
			}
		}
	}

	return nil
}

// collectSidechainRoots collects all sidechain root entries
func (s *SidechainProcessor) collectSidechainRoots(entries []models.LogEntry, entryMap map[string]*models.ProcessedEntry) []*models.ProcessedEntry {
	var sidechainRoots []*models.ProcessedEntry

	for _, entry := range entries {
		processed := entryMap[entry.UUID]
		if processed.IsSidechain && processed.ParentUUID == "" && !processed.IsToolResult {
			sidechainRoots = append(sidechainRoots, processed)
		}
	}

	return sidechainRoots
}

// matchTaskWithSidechain matches a Task tool call with its corresponding sidechain conversation
func (s *SidechainProcessor) matchTaskWithSidechain(
	toolCall *models.ToolCall,
	entry *models.LogEntry,
	originalEntries []models.LogEntry,
	sidechainRoots []*models.ProcessedEntry,
	entryMap map[string]*models.ProcessedEntry,
	matchedSidechains map[string]bool,
) {
	if debug.Enabled {
		log.Printf("Found Task tool %s in entry %s (sidechain: %v)",
			toolCall.ID, entry.UUID, entry.IsSidechain)
	}

	// Extract the prompt from the Task tool call
	taskPrompt := s.extractTaskPrompt(toolCall)
	if taskPrompt == "" {
		if debug.Enabled {
			log.Printf("Task tool %s has empty prompt, skipping", toolCall.ID)
		}
		return
	}

	if debug.Enabled {
		log.Printf("Task tool %s prompt: %.50s...", toolCall.ID, taskPrompt)
	}

	// Extract the result text from the tool result
	taskResult := s.extractTaskResult(toolCall, originalEntries)
	if taskResult == "" {
		if debug.Enabled {
			log.Printf("Task tool %s has empty result, skipping", toolCall.ID)
		}
		return
	}

	if debug.Enabled {
		log.Printf("Task tool %s result: %.50s...", toolCall.ID, taskResult)
	}

	// Find the best matching sidechain
	bestMatch, bestMatchScore := s.findBestMatchingSidechain(
		toolCall, taskPrompt, taskResult, sidechainRoots, entryMap, matchedSidechains,
	)

	if bestMatch != nil {
		toolCall.TaskEntries = s.collectSidechainEntries(bestMatch, entryMap)
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

// extractTaskPrompt extracts the prompt from a Task tool call's raw input
func (s *SidechainProcessor) extractTaskPrompt(toolCall *models.ToolCall) string {
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

// extractTaskResult extracts the result text from a Task tool's result
func (s *SidechainProcessor) extractTaskResult(toolCall *models.ToolCall, originalEntries []models.LogEntry) string {
	if toolCall.Result == nil {
		return ""
	}

	// Find the original entry for this result
	for _, e := range originalEntries {
		if e.UUID == toolCall.Result.UUID {
			var msg map[string]interface{}
			if err := json.Unmarshal(e.Message, &msg); err == nil {
				if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
					if toolResult, ok := content[0].(map[string]interface{}); ok {
						if resultContent, ok := toolResult["content"].([]interface{}); ok && len(resultContent) > 0 {
							if textContent, ok := resultContent[0].(map[string]interface{}); ok {
								if text, ok := textContent["text"].(string); ok {
									return text
								}
							}
						}
					}
				}
			}
			break
		}
	}

	return ""
}

// findBestMatchingSidechain finds the best matching sidechain for a Task tool
func (s *SidechainProcessor) findBestMatchingSidechain(
	toolCall *models.ToolCall,
	taskPrompt, taskResult string,
	sidechainRoots []*models.ProcessedEntry,
	entryMap map[string]*models.ProcessedEntry,
	matchedSidechains map[string]bool,
) (*models.ProcessedEntry, int) {
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
		firstUser := s.getFirstUserMessage(sidechain, entryMap)
		lastAssistant := s.getLastAssistantMessage(sidechain, entryMap)

		if firstUser == "" || lastAssistant == "" {
			if debug.Enabled {
				log.Printf("Sidechain %s has empty first user or last assistant, skipping",
					sidechain.UUID)
			}
			continue
		}

		// Calculate match score
		score := s.calculateMatchScore(toolCall, taskPrompt, taskResult, firstUser, lastAssistant, sidechain)

		if score > bestMatchScore {
			bestMatchScore = score
			bestMatch = sidechain
		}

		// If we have a perfect match (both prompt and result), we can stop looking
		if score == 2 {
			break
		}
	}

	return bestMatch, bestMatchScore
}

// calculateMatchScore calculates the match score between a Task tool and a sidechain
func (s *SidechainProcessor) calculateMatchScore(
	toolCall *models.ToolCall,
	taskPrompt, taskResult, firstUser, lastAssistant string,
	sidechain *models.ProcessedEntry,
) int {
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

	return score
}

// collectSidechainEntries collects all entries in a sidechain conversation
func (s *SidechainProcessor) collectSidechainEntries(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) []*models.ProcessedEntry {
	return collectSidechainEntries(root, entryMap)
}

// getFirstUserMessage finds the first user message in a sidechain conversation
func (s *SidechainProcessor) getFirstUserMessage(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	return getFirstUserMessage(root, entryMap)
}

// getLastAssistantMessage finds the last assistant message in a sidechain conversation
func (s *SidechainProcessor) getLastAssistantMessage(root *models.ProcessedEntry, entryMap map[string]*models.ProcessedEntry) string {
	return getLastAssistantMessage(root, entryMap)
}

// identifySidechainBoundaries identifies the start and end points of sidechain conversations
func (s *SidechainProcessor) identifySidechainBoundaries(entries []*models.ProcessedEntry) []SidechainContext {
	var contexts []SidechainContext

	// This functionality is now handled by collecting sidechain roots and building trees
	// The boundaries are implicit in the parent-child relationships

	return contexts
}

// groupSidechainEntries creates a grouped entry for a sidechain conversation
func (s *SidechainProcessor) groupSidechainEntries(ctx SidechainContext, entries []*models.ProcessedEntry) *models.ProcessedEntry {
	// This functionality is now handled by collectSidechainEntries
	// which builds the tree structure for display
	return nil
}
