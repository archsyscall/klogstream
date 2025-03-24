package klogstream

import (
	"regexp"
	"time"

	"github.com/archsyscall/klogstream/internal/filter"
	"k8s.io/apimachinery/pkg/labels"
)

// LogFilter defines filtering criteria for kubernetes logs
type LogFilter struct {
	// PodNameRegex filters pods by name regex
	PodNameRegex *regexp.Regexp
	// ContainerRegex filters containers by name regex
	ContainerRegex *regexp.Regexp
	// LabelSelector filters pods by their labels
	LabelSelector labels.Selector
	// IncludeRegex only includes log lines matching this regex
	IncludeRegex *regexp.Regexp
	// Since only includes logs newer than this time
	Since *time.Time
	// ContainerState filters by container state ("all", "running", "terminated", ...)
	ContainerState string
	// Namespaces is a list of namespaces to filter logs from
	Namespaces []string
}

// NewLogFilterBuilder creates a new LogFilterBuilder
func NewLogFilterBuilder() *LogFilterBuilder {
	return &LogFilterBuilder{
		builder: filter.NewLogFilterBuilder(),
	}
}

// LogFilterBuilder provides a fluent API for building LogFilter
type LogFilterBuilder struct {
	builder *filter.LogFilterBuilder
}

// PodRegex sets the pod name regex pattern
func (b *LogFilterBuilder) PodRegex(pattern string) *LogFilterBuilder {
	b.builder.PodRegex(pattern)
	return b
}

// ContainerRegex sets the container name regex pattern
func (b *LogFilterBuilder) ContainerRegex(pattern string) *LogFilterBuilder {
	b.builder.ContainerRegex(pattern)
	return b
}

// Label adds a label selector
func (b *LogFilterBuilder) Label(key, value string) *LogFilterBuilder {
	b.builder.Label(key, value)
	return b
}

// Include sets the regex for log lines to include
func (b *LogFilterBuilder) Include(pattern string) *LogFilterBuilder {
	b.builder.Include(pattern)
	return b
}

// Since sets the time to stream logs from
func (b *LogFilterBuilder) Since(duration time.Duration) *LogFilterBuilder {
	b.builder.Since(duration)
	return b
}

// ContainerState sets the container state filter
func (b *LogFilterBuilder) ContainerState(state string) *LogFilterBuilder {
	b.builder.ContainerState(state)
	return b
}

// Namespace adds a namespace to filter
func (b *LogFilterBuilder) Namespace(namespace string) *LogFilterBuilder {
	b.builder.Namespace(namespace)
	return b
}

// Build creates and validates the LogFilter
func (b *LogFilterBuilder) Build() (*LogFilter, error) {
	internalFilter, err := b.builder.Build()
	if err != nil {
		return nil, err
	}

	return &LogFilter{
		PodNameRegex:   internalFilter.PodNameRegex,
		ContainerRegex: internalFilter.ContainerRegex,
		LabelSelector:  internalFilter.LabelSelector,
		IncludeRegex:   internalFilter.IncludeRegex,
		Since:          internalFilter.Since,
		ContainerState: internalFilter.ContainerState,
		Namespaces:     internalFilter.Namespaces,
	}, nil
}
