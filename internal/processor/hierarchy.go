package processor

import (
	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/models"
)

// HierarchyBuilder builds conversation hierarchy and calculates depths.
type HierarchyBuilder struct{}

// NewHierarchyBuilder creates a new hierarchy builder
func NewHierarchyBuilder() *HierarchyBuilder {
	return &HierarchyBuilder{}
}

// BuildHierarchy builds the hierarchy and sets depths for all entries
func (h *HierarchyBuilder) BuildHierarchy(entries []*models.ProcessedEntry) error {
	// Set depth for all entries based on sidechain hierarchy
	// Root conversation starts at depth 1
	for _, entry := range entries {
		h.setEntryDepth(entry, constants.RootConversationDepth)
	}

	return nil
}

// setEntryDepth recursively sets the depth for entries based on sidechain hierarchy
func (h *HierarchyBuilder) setEntryDepth(entry *models.ProcessedEntry, depth int) {
	// Set the depth for this entry
	entry.Depth = depth

	// Process all tool calls
	for i := range entry.ToolCalls {
		toolCall := &entry.ToolCalls[i]

		// If this is a Task tool with sidechain entries, set their depth to current depth + 1
		if toolCall.Name == constants.TaskToolName && len(toolCall.TaskEntries) > 0 {
			for _, taskEntry := range toolCall.TaskEntries {
				h.setEntryDepth(taskEntry, depth+1)
			}
		}

		// Also set depth for tool results
		if toolCall.Result != nil {
			toolCall.Result.Depth = depth
		}
	}

	// Process children (though main conversation entries shouldn't have children)
	for _, child := range entry.Children {
		h.setEntryDepth(child, depth)
	}
}

