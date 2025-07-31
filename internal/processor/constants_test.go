package processor

import "testing"

func TestConstants(t *testing.T) {
	// Test tool name constants
	toolNames := map[string]string{
		"Task":      ToolNameTask,
		"Bash":      ToolNameBash,
		"Edit":      ToolNameEdit,
		"MultiEdit": ToolNameMultiEdit,
		"Write":     ToolNameWrite,
		"Read":      ToolNameRead,
		"TodoWrite": ToolNameTodoWrite,
	}
	
	for expected, actual := range toolNames {
		if actual != expected {
			t.Errorf("Expected tool name %s, got %s", expected, actual)
		}
	}
	
	// Test entry type constants
	entryTypes := map[string]string{
		"user":        TypeUser,
		"assistant":   TypeAssistant,
		"message":     TypeMessage,
		"tool_use":    TypeToolUse,
		"tool_result": TypeToolResult,
	}
	
	for expected, actual := range entryTypes {
		if actual != expected {
			t.Errorf("Expected entry type %s, got %s", expected, actual)
		}
	}
	
	// Test role constants
	if RoleUser != "user" {
		t.Errorf("Expected RoleUser='user', got %s", RoleUser)
	}
	if RoleAssistant != "assistant" {
		t.Errorf("Expected RoleAssistant='assistant', got %s", RoleAssistant)
	}
}