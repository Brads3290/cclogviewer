package processor

import (
	"log"

	"github.com/brads3290/cclogviewer/internal/models"
)

// initializeProcessingState creates and initializes a new processing state
func initializeProcessingState(capacity int) *ProcessingState {
	return &ProcessingState{
		Entries:        make([]*models.ProcessedEntry, 0, capacity),
		ToolCallMap:    make(map[string]*ToolCallContext),
		ParentChildMap: make(map[string][]string),
	}
}

// processAllEntries processes raw log entries into ProcessedEntry objects
func processAllEntries(entries []models.LogEntry, state *ProcessingState, entryMap map[string]*models.ProcessedEntry) {
	for i, entry := range entries {
		state.Index = i
		processed := processEntry(entry)
		entryMap[processed.UUID] = processed
		state.Entries = append(state.Entries, processed)
	}
}

// matchToolCallsWithResults matches tool calls with their corresponding results
func matchToolCallsWithResults(entries []*models.ProcessedEntry) {
	matcher := NewToolCallMatcher()
	if err := matcher.MatchToolCalls(entries); err != nil {
		log.Printf("Error matching tool calls: %v", err)
	}
}

// processSidechainConversations processes Task tool sidechain conversations
func processSidechainConversations(state *ProcessingState, entries []models.LogEntry, entryMap map[string]*models.ProcessedEntry) {
	sidechainProc := NewSidechainProcessor()
	if err := sidechainProc.ProcessSidechains(state.Entries, entries, entryMap); err != nil {
		log.Printf("Error processing sidechains: %v", err)
	}
}

// getRootEntries filters entries to get only root-level entries
func getRootEntries(state *ProcessingState) []*models.ProcessedEntry {
	matcher := NewToolCallMatcher()
	return matcher.FilterRootEntries(state.Entries)
}

// calculateAllTokens calculates token counts for all entries
func calculateAllTokens(rootEntries []*models.ProcessedEntry) {
	for _, entry := range rootEntries {
		calculateTokensForEntry(entry)
	}
}

// checkAllMissingResults checks for missing tool results across all entries
func checkAllMissingResults(rootEntries []*models.ProcessedEntry) {
	for _, entry := range rootEntries {
		checkMissingToolResults(entry)
	}
}

// linkAllCommandOutputs links command messages with their outputs
func linkAllCommandOutputs(rootEntries []*models.ProcessedEntry) {
	linkCommandOutputs(rootEntries)
}

// buildFinalHierarchy builds the final hierarchy and sets depths
func buildFinalHierarchy(rootEntries []*models.ProcessedEntry) {
	hierarchy := NewHierarchyBuilder()
	if err := hierarchy.BuildHierarchy(rootEntries); err != nil {
		log.Printf("Error building hierarchy: %v", err)
	}
}
