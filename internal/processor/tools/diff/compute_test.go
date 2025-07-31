package diff_test

import (
	"testing"

	"github.com/Brads3290/cclogviewer/internal/processor/tools/diff"
)

func TestComputeLineDiff(t *testing.T) {
	tests := []struct {
		name     string
		oldStr   string
		newStr   string
		expected []diff.LineType
	}{
		{
			name:   "identical strings",
			oldStr: "line1\nline2\nline3",
			newStr: "line1\nline2\nline3",
			expected: []diff.LineType{
				diff.LineUnchanged,
				diff.LineUnchanged,
				diff.LineUnchanged,
			},
		},
		{
			name:   "line modified",
			oldStr: "line1\nline2\nline3",
			newStr: "line1\nline2-modified\nline3",
			expected: []diff.LineType{
				diff.LineUnchanged,
				diff.LineRemoved,
				diff.LineAdded,
				diff.LineUnchanged,
			},
		},
		{
			name:   "line added",
			oldStr: "line1\nline3",
			newStr: "line1\nline2\nline3",
			expected: []diff.LineType{
				diff.LineUnchanged,
				diff.LineAdded,
				diff.LineUnchanged,
			},
		},
		{
			name:   "line removed",
			oldStr: "line1\nline2\nline3",
			newStr: "line1\nline3",
			expected: []diff.LineType{
				diff.LineUnchanged,
				diff.LineRemoved,
				diff.LineUnchanged,
			},
		},
		{
			name:   "complete replacement",
			oldStr: "old1\nold2",
			newStr: "new1\nnew2",
			expected: []diff.LineType{
				diff.LineRemoved,
				diff.LineRemoved,
				diff.LineAdded,
				diff.LineAdded,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diff.ComputeLineDiff(tt.oldStr, tt.newStr)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d diff lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, line := range result {
				if line.Type != tt.expected[i] {
					t.Errorf("Line %d: expected type %v, got %v", i, tt.expected[i], line.Type)
				}
			}
		})
	}
}

func TestLineType(t *testing.T) {
	tests := []struct {
		lineType diff.LineType
		str      string
		prefix   string
		cssClass string
	}{
		{
			lineType: diff.LineUnchanged,
			str:      "unchanged",
			prefix:   " ",
			cssClass: "line-unchanged",
		},
		{
			lineType: diff.LineAdded,
			str:      "added",
			prefix:   "+",
			cssClass: "line-added",
		},
		{
			lineType: diff.LineRemoved,
			str:      "removed",
			prefix:   "-",
			cssClass: "line-removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if tt.lineType.String() != tt.str {
				t.Errorf("Expected string %s, got %s", tt.str, tt.lineType.String())
			}
			if tt.lineType.Prefix() != tt.prefix {
				t.Errorf("Expected prefix %s, got %s", tt.prefix, tt.lineType.Prefix())
			}
			if tt.lineType.CSSClass() != tt.cssClass {
				t.Errorf("Expected CSS class %s, got %s", tt.cssClass, tt.lineType.CSSClass())
			}
		})
	}
}
