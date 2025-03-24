package filter

import (
	"regexp"
	"time"

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

// DefaultContainerState is the default container state to filter by
const DefaultContainerState = "all"

// NewLogFilter creates a new LogFilter with default values
func NewLogFilter() *LogFilter {
	return &LogFilter{
		ContainerState: DefaultContainerState,
	}
}

// IsEmpty returns true if no filter criteria are set
func (f *LogFilter) IsEmpty() bool {
	return f.PodNameRegex == nil &&
		f.ContainerRegex == nil &&
		f.LabelSelector == nil &&
		f.IncludeRegex == nil &&
		f.Since == nil &&
		(f.ContainerState == DefaultContainerState || f.ContainerState == "") &&
		len(f.Namespaces) == 0
}

// Validate checks if the filter is valid
func (f *LogFilter) Validate() error {
	if f.IsEmpty() {
		return ErrEmptyFilter
	}

	if len(f.Namespaces) == 0 {
		return ErrNoNamespaceSpecified
	}

	if f.ContainerState != "" &&
		f.ContainerState != "all" &&
		f.ContainerState != "running" &&
		f.ContainerState != "terminated" {
		return ErrInvalidContainerState
	}

	if f.Since != nil && f.Since.After(time.Now()) {
		return ErrInvalidSinceTime
	}

	return nil
}
