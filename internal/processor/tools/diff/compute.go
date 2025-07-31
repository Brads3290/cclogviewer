package diff

import (
	"strings"
)

// isMatchingLCSPosition checks if current positions match in the LCS
func isMatchingLCSPosition(lcsIdx, oldIdx, newIdx int, lcs, oldLines, newLines []string) bool {
	if lcsIdx >= len(lcs) || oldIdx >= len(oldLines) || newIdx >= len(newLines) {
		return false
	}
	return oldLines[oldIdx] == lcs[lcsIdx] && newLines[newIdx] == lcs[lcsIdx]
}

// ComputeLineDiff computes a simple line-based diff between two strings
func ComputeLineDiff(oldStr, newStr string) []DiffLine {
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	// If strings are identical, return all unchanged lines
	if oldStr == newStr {
		diff := make([]DiffLine, len(oldLines))
		for i, line := range oldLines {
			diff[i] = DiffLine{
				Type:    LineUnchanged,
				Content: line,
				LineNum: i + 1,
			}
		}
		return diff
	}

	// Simple diff: find longest common subsequence
	lcs := longestCommonSubsequence(oldLines, newLines)

	// Build diff from LCS
	diff := []DiffLine{}
	oldIdx, newIdx := 0, 0
	lcsIdx := 0
	lineNum := 1

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if isMatchingLCSPosition(lcsIdx, oldIdx, newIdx, lcs, oldLines, newLines) {
			// Common line
			diff = append(diff, DiffLine{
				Type:    LineUnchanged,
				Content: oldLines[oldIdx],
				LineNum: lineNum,
			})
			oldIdx++
			newIdx++
			lcsIdx++
			lineNum++
		} else if oldIdx < len(oldLines) && (lcsIdx >= len(lcs) || oldLines[oldIdx] != lcs[lcsIdx]) {
			// Removed line
			diff = append(diff, DiffLine{
				Type:    LineRemoved,
				Content: oldLines[oldIdx],
				LineNum: lineNum,
			})
			oldIdx++
			lineNum++
		} else if newIdx < len(newLines) && (lcsIdx >= len(lcs) || newLines[newIdx] != lcs[lcsIdx]) {
			// Added line
			diff = append(diff, DiffLine{
				Type:    LineAdded,
				Content: newLines[newIdx],
				LineNum: lineNum,
			})
			newIdx++
			lineNum++
		}
	}

	return diff
}

// longestCommonSubsequence finds the LCS of two string slices
func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// Build the DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Reconstruct the LCS
	lcs := []string{}
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

// ComputeUnifiedDiff generates a unified diff format string
func ComputeUnifiedDiff(oldStr, newStr string, contextLines int) string {
	lines := ComputeLineDiff(oldStr, newStr)

	var result strings.Builder

	for _, line := range lines {
		result.WriteString(line.Type.Prefix())
		result.WriteString(line.Content)
		result.WriteString("\n")
	}

	return result.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
