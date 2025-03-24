package matcher

import (
	"regexp"
	"strings"
)

// JavaStackMatcher detects Java stack traces for multiline log merging
type JavaStackMatcher struct {
	// ExceptionRegex matches Java exception lines
	ExceptionRegex *regexp.Regexp
	// TabIndentationRegex matches lines with tab indentation (common in stack traces)
	TabIndentationRegex *regexp.Regexp
	// CausedByRegex matches "Caused by:" lines in stack traces
	CausedByRegex *regexp.Regexp
	// AtRegex matches "at " lines in stack traces
	AtRegex *regexp.Regexp
}

// NewJavaStackMatcher creates a new JavaStackMatcher
func NewJavaStackMatcher() *JavaStackMatcher {
	return &JavaStackMatcher{
		ExceptionRegex:      regexp.MustCompile(`^[^\s]+Exception:`),
		TabIndentationRegex: regexp.MustCompile(`^\t`),
		CausedByRegex:       regexp.MustCompile(`^Caused by:`),
		AtRegex:             regexp.MustCompile(`^\s+at `),
	}
}

// ShouldMerge determines if the next line should be merged with the previous line
func (m *JavaStackMatcher) ShouldMerge(previous, next string) bool {
	// If the line is empty, don't merge
	if strings.TrimSpace(next) == "" {
		return false
	}

	// Check for tab indentation (common in stack traces)
	if m.TabIndentationRegex.MatchString(next) {
		return true
	}

	// Check for "at " lines (common in Java stack traces)
	if m.AtRegex.MatchString(next) {
		return true
	}

	// Check for "Caused by:" lines
	if m.CausedByRegex.MatchString(next) {
		return true
	}

	// If the previous line ends with a backslash, merge
	if strings.HasSuffix(strings.TrimSpace(previous), "\\") {
		return true
	}

	return false
}
