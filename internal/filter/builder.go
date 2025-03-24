package filter

import (
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/labels"
)

// LogFilterBuilder provides a fluent API for building LogFilter
type LogFilterBuilder struct {
	filter *LogFilter
}

// NewLogFilterBuilder creates a new LogFilterBuilder
func NewLogFilterBuilder() *LogFilterBuilder {
	return &LogFilterBuilder{
		filter: NewLogFilter(),
	}
}

// PodRegex sets the pod name regex pattern
func (b *LogFilterBuilder) PodRegex(pattern string) *LogFilterBuilder {
	if pattern != "" {
		regex, err := regexp.Compile(pattern)
		if err == nil {
			b.filter.PodNameRegex = regex
		}
	}
	return b
}

// ContainerRegex sets the container name regex pattern
func (b *LogFilterBuilder) ContainerRegex(pattern string) *LogFilterBuilder {
	if pattern != "" {
		regex, err := regexp.Compile(pattern)
		if err == nil {
			b.filter.ContainerRegex = regex
		}
	}
	return b
}

// Label adds a label selector
func (b *LogFilterBuilder) Label(key, value string) *LogFilterBuilder {
	if key != "" {
		sel := labels.SelectorFromSet(labels.Set{key: value})
		b.filter.LabelSelector = sel
	}
	return b
}

// Include sets the regex for log lines to include
func (b *LogFilterBuilder) Include(pattern string) *LogFilterBuilder {
	if pattern != "" {
		regex, err := regexp.Compile(pattern)
		if err == nil {
			b.filter.IncludeRegex = regex
		}
	}
	return b
}

// Since sets the time to stream logs from
func (b *LogFilterBuilder) Since(duration time.Duration) *LogFilterBuilder {
	if duration >= 0 {
		tm := time.Now().Add(-duration)
		b.filter.Since = &tm
	}
	return b
}

// ContainerState sets the container state filter
func (b *LogFilterBuilder) ContainerState(state string) *LogFilterBuilder {
	if state != "" {
		b.filter.ContainerState = state
	}
	return b
}

// Namespace adds a namespace to filter
func (b *LogFilterBuilder) Namespace(namespace string) *LogFilterBuilder {
	if namespace != "" {
		b.filter.Namespaces = append(b.filter.Namespaces, namespace)
	}
	return b
}

// Build creates and validates the LogFilter
func (b *LogFilterBuilder) Build() (*LogFilter, error) {
	err := b.filter.Validate()
	if err != nil {
		return nil, err
	}
	return b.filter, nil
}
