package processor

import "github.com/Brads3290/cclogviewer/internal/models"

// HierarchyBuilder builds parent-child relationships and calculates depths
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
		h.setEntryDepth(entry, 1)
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
		if toolCall.Name == ToolNameTask && len(toolCall.TaskEntries) > 0 {
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

// calculateDepths calculates depths for all entries based on their relationships
func (h *HierarchyBuilder) calculateDepths(entries []*models.ProcessedEntry, parentChildMap map[string][]string) {
	// This is now handled by setEntryDepth which is more appropriate for our use case
	// since depth is based on sidechain nesting, not parent-child relationships
}

// findRootEntries identifies root-level entries (entries without parents)
func (h *HierarchyBuilder) findRootEntries(entries []*models.ProcessedEntry) []*models.ProcessedEntry {
	var roots []*models.ProcessedEntry
	for _, entry := range entries {
		if entry.ParentUUID == "" {
			roots = append(roots, entry)
		}
	}
	return roots
}

// BuildParentChildMap builds a map of parent UUIDs to child UUIDs
func (h *HierarchyBuilder) BuildParentChildMap(entries []*models.ProcessedEntry) map[string][]string {
	parentChildMap := make(map[string][]string)

	for _, entry := range entries {
		if entry.ParentUUID != "" {
			parentChildMap[entry.ParentUUID] = append(parentChildMap[entry.ParentUUID], entry.UUID)
		}
	}

	return parentChildMap
}
