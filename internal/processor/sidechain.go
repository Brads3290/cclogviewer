package processor

import (
	"encoding/json"
	"fmt"
	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/debug"
	"github.com/brads3290/cclogviewer/internal/models"
	"log"
	"strings"
)

// SidechainProcessor handles Task tool sidechain conversation processing.
type SidechainProcessor struct{}

// NewSidechainProcessor creates a new sidechain processor
func NewSidechainProcessor() *SidechainProcessor {
	return &SidechainProcessor{}
}

// TaskMatchContext holds context for Task tool sidechain matching.
type TaskMatchContext struct {
	ToolCall          *models.ToolCall
	Entry             *models.LogEntry
	OriginalEntries   []models.LogEntry
	SidechainRoots    []*models.ProcessedEntry
	EntryMap          map[string]*models.ProcessedEntry
	MatchedSidechains map[string]bool
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
		if entry.Type == constants.TypeAssistant {
			processed := entryMap[entry.UUID]

			if debug.Enabled && len(processed.ToolCalls) > 0 {
				log.Printf("Processing assistant entry %s (sidechain: %v) with %d tool calls",
					entry.UUID, entry.IsSidechain, len(processed.ToolCalls))
			}

			for i := range processed.ToolCalls {
				toolCall := &processed.ToolCalls[i]
				if toolCall.Name == constants.TaskToolName {
					ctx := &TaskMatchContext{
						ToolCall:          toolCall,
						Entry:             &entry,
						OriginalEntries:   originalEntries,
						SidechainRoots:    sidechainRoots,
						EntryMap:          entryMap,
						MatchedSidechains: matchedSidechains,
					}
					s.matchTaskWithSidechain(ctx)
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
		if s.isSidechainRoot(processed) {
			sidechainRoots = append(sidechainRoots, processed)
		}
	}

	return sidechainRoots
}

// isSidechainRoot checks if an entry is a sidechain root
func (s *SidechainProcessor) isSidechainRoot(processed *models.ProcessedEntry) bool {
	return processed.IsSidechain && processed.ParentUUID == "" && !processed.IsToolResult
}

// matchTaskWithSidechain matches a Task tool call with its corresponding sidechain conversation
func (s *SidechainProcessor) matchTaskWithSidechain(ctx *TaskMatchContext) {
	if debug.Enabled {
		log.Printf("Found Task tool %s in entry %s (sidechain: %v)",
			ctx.ToolCall.ID, ctx.Entry.UUID, ctx.Entry.IsSidechain)
	}

	// Extract the prompt from the Task tool call
	taskPrompt := s.extractTaskPrompt(ctx.ToolCall)
	if taskPrompt == "" {
		if debug.Enabled {
			log.Printf("Task tool %s has empty prompt, skipping", ctx.ToolCall.ID)
		}
		return
	}

	if debug.Enabled {
		log.Printf("Task tool %s prompt: %.50s...", ctx.ToolCall.ID, taskPrompt)
	}

	// Extract the result text from the tool result
	taskResult := s.extractTaskResult(ctx.ToolCall, ctx.OriginalEntries)
	if taskResult == "" {
		if debug.Enabled {
			log.Printf("Task tool %s has empty result, skipping", ctx.ToolCall.ID)
		}
		return
	}

	if debug.Enabled {
		log.Printf("Task tool %s result: %.50s...", ctx.ToolCall.ID, taskResult)
	}

	// Find the best matching sidechain
	bestMatch, bestMatchScore := s.findBestMatchingSidechain(
		ctx.ToolCall, taskPrompt, taskResult, ctx.SidechainRoots, ctx.EntryMap, ctx.MatchedSidechains,
	)

	if bestMatch != nil {
		ctx.ToolCall.TaskEntries = collectSidechainEntries(bestMatch, ctx.EntryMap)
		ctx.MatchedSidechains[bestMatch.UUID] = true
		if debug.Enabled {
			log.Printf("Matched Task tool %s to sidechain %s (score: %d, entries: %d)",
				ctx.ToolCall.ID, bestMatch.UUID, bestMatchScore, len(ctx.ToolCall.TaskEntries))
		}
	} else {
		if debug.Enabled {
			log.Printf("No match found for Task tool %s", ctx.ToolCall.ID)
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
		if e.UUID != toolCall.Result.UUID {
			continue
		}

		text, err := s.extractTextFromEntry(e)
		if err == nil {
			return text
		}
		break
	}

	return ""
}

// extractTextFromEntry extracts text content from a log entry
func (s *SidechainProcessor) extractTextFromEntry(entry models.LogEntry) (string, error) {
	var msg map[string]interface{}
	if err := json.Unmarshal(entry.Message, &msg); err != nil {
		return "", err
	}

	content, ok := s.getContentArray(msg)
	if !ok {
		return "", fmt.Errorf("no content array")
	}

	return s.extractTextFromContent(content)
}

// getContentArray safely extracts the content array from a message
func (s *SidechainProcessor) getContentArray(msg map[string]interface{}) ([]interface{}, bool) {
	content, ok := msg["content"].([]interface{})
	if !ok || len(content) == 0 {
		return nil, false
	}
	return content, true
}

// extractTextFromContent extracts text from content array
func (s *SidechainProcessor) extractTextFromContent(content []interface{}) (string, error) {
	toolResult, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid tool result format")
	}

	resultContent, ok := toolResult["content"].([]interface{})
	if !ok || len(resultContent) == 0 {
		return "", fmt.Errorf("no result content")
	}

	textContent, ok := resultContent[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid text content format")
	}

	text, ok := textContent["text"].(string)
	if !ok {
		return "", fmt.Errorf("no text field")
	}

	return text, nil
}

// canCheckPrefixMatch checks if both strings are long enough for prefix matching
func (s *SidechainProcessor) canCheckPrefixMatch(str1, str2 string) bool {
	return len(str1) > constants.MinTextLengthForPrefixMatch && len(str2) > constants.MinTextLengthForPrefixMatch
}

// hasPrefixMatch checks if either string is a prefix of the other
func (s *SidechainProcessor) hasPrefixMatch(str1, str2 string) bool {
	return strings.HasPrefix(str1, str2) || strings.HasPrefix(str2, str1)
}

// findBestMatchingSidechain scores matches to handle concurrent Task invocations.
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
		firstUser := getFirstUserMessage(sidechain, entryMap)
		lastAssistant := getLastAssistantMessage(sidechain, entryMap)

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
		if score == constants.PerfectMatchScore {
			break
		}
	}

	return bestMatch, bestMatchScore
}

// calculateMatchScore uses prompt and result matching to disambiguate sidechains.
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
	if !promptMatch && s.canCheckPrefixMatch(taskPromptNorm, firstUserNorm) {
		promptMatch = s.hasPrefixMatch(taskPromptNorm, firstUserNorm)
	}
	if !resultMatch && s.canCheckPrefixMatch(taskResultNorm, lastAssistantNorm) {
		resultMatch = s.hasPrefixMatch(taskResultNorm, lastAssistantNorm)
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
