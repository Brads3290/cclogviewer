package diff

// LineType represents the type of change in a diff line
type LineType int

const (
	// LineUnchanged represents an unchanged line
	LineUnchanged LineType = iota
	// LineAdded represents an added line
	LineAdded
	// LineRemoved represents a removed line
	LineRemoved
)

// DiffLine represents a line in the diff with its type and content
type DiffLine struct {
	Type    LineType
	Content string
	LineNum int
}

// String returns the string representation of the line type
func (lt LineType) String() string {
	switch lt {
	case LineUnchanged:
		return "unchanged"
	case LineAdded:
		return "added"
	case LineRemoved:
		return "removed"
	default:
		return "unknown"
	}
}

// Prefix returns the diff prefix character for the line type
func (lt LineType) Prefix() string {
	switch lt {
	case LineUnchanged:
		return " "
	case LineAdded:
		return "+"
	case LineRemoved:
		return "-"
	default:
		return "?"
	}
}

// CSSClass returns the CSS class name for the line type
func (lt LineType) CSSClass() string {
	switch lt {
	case LineUnchanged:
		return "line-unchanged"
	case LineAdded:
		return "line-added"
	case LineRemoved:
		return "line-removed"
	default:
		return ""
	}
}
