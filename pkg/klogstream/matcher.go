package klogstream

import (
	"github.com/archsyscall/klogstream/internal/matcher"
)

// JavaStackMatcher detects Java stack traces for multiline log merging
type JavaStackMatcher struct {
	internal *matcher.JavaStackMatcher
}

// NewJavaStackMatcher creates a new JavaStackMatcher
func NewJavaStackMatcher() *JavaStackMatcher {
	return &JavaStackMatcher{
		internal: matcher.NewJavaStackMatcher(),
	}
}

// ShouldMerge determines if the next line should be merged with the previous line
func (m *JavaStackMatcher) ShouldMerge(previous, next string) bool {
	return m.internal.ShouldMerge(previous, next)
}

// JSONMatcher detects JSON formatted logs for multiline log merging
type JSONMatcher struct {
	internal *matcher.JSONMatcher
}

// NewJSONMatcher creates a new JSONMatcher
func NewJSONMatcher() *JSONMatcher {
	return &JSONMatcher{
		internal: matcher.NewJSONMatcher(),
	}
}

// ShouldMerge determines if the next line should be merged with the previous line
func (m *JSONMatcher) ShouldMerge(previous, next string) bool {
	return m.internal.ShouldMerge(previous, next)
}
