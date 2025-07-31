package processor

// Tool names
const (
	ToolNameTask      = "Task"
	ToolNameBash      = "Bash"
	ToolNameEdit      = "Edit"
	ToolNameMultiEdit = "MultiEdit"
	ToolNameWrite     = "Write"
	ToolNameRead      = "Read"
	ToolNameTodoWrite = "TodoWrite"
)

// Entry types
const (
	TypeUser       = "user"
	TypeAssistant  = "assistant"
	TypeMessage    = "message"
	TypeToolUse    = "tool_use"
	TypeToolResult = "tool_result"
)

// Roles
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Content types
const (
	ContentTypeText       = "text"
	ContentTypeToolUse    = "tool_use"
	ContentTypeToolResult = "tool_result"
)

// XML tags for command parsing
const (
	TagCommandName   = "command-name"
	TagCommandArgs   = "command-args"
	TagCommandStdout = "local-command-stdout"
)

// UI and formatting constants
const (
	BashOutputCollapseThreshold = 20
	ShortUUIDLength             = 8
	MinTextLengthForPrefixMatch = 20
)

// Parser constants
const (
	DefaultBufferSize = 64 * 1024        // 64KB buffer
	MaxLineSize       = 10 * 1024 * 1024 // 10MB max line size
)

// Number formatting constants
const (
	ThousandSeparatorThreshold = 1000
)

// Builder constants
const (
	HTMLBuilderInitialCapacity = 100
)
