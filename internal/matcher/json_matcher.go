package matcher

import (
	"regexp"
	"strings"
)

// JSONMatcher detects JSON formatted logs for multiline log merging
type JSONMatcher struct {
	// StartObjectRegex matches the start of a JSON object
	StartObjectRegex *regexp.Regexp
	// EndObjectRegex matches the end of a JSON object
	EndObjectRegex *regexp.Regexp
	// BracketMode tracks open/close brackets for proper JSON structure
	BracketMode bool
	// BracketCount keeps track of open/close brackets
	BracketCount int
}

// NewJSONMatcher creates a new JSONMatcher
func NewJSONMatcher() *JSONMatcher {
	return &JSONMatcher{
		StartObjectRegex: regexp.MustCompile(`^[^{]*{`),
		EndObjectRegex:   regexp.MustCompile(`}[^}]*$`),
		BracketMode:      true,
	}
}

// ShouldMerge determines if the next line should be merged with the previous line
func (m *JSONMatcher) ShouldMerge(previous, next string) bool {
	// Check for continuation with backslash
	if strings.HasSuffix(strings.TrimSpace(previous), "\\") {
		return true
	}

	// If not in bracket mode, use simple heuristics
	if !m.BracketMode {
		return false
	}

	// Reset bracket count when encountering a new JSON object start
	if m.StartObjectRegex.MatchString(previous) && m.BracketCount == 0 {
		m.BracketCount = 1
		// Count additional open brackets
		for _, c := range previous {
			if c == '{' {
				m.BracketCount++
			} else if c == '}' {
				m.BracketCount--
			}
		}
		// Adjust for the initial match we already counted
		m.BracketCount--
	} else {
		// Update bracket count based on previous line
		for _, c := range previous {
			if c == '{' {
				m.BracketCount++
			} else if c == '}' {
				m.BracketCount--
			}
		}
	}

	// Don't continue if bracket count is zero or negative
	if m.BracketCount <= 0 {
		m.BracketCount = 0 // Reset to avoid negative counts
		return false
	}

	return true
}
