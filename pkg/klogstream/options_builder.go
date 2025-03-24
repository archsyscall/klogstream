package klogstream

import (
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/labels"
)

// WithNamespace adds a namespace to the log filter
func WithNamespace(namespace string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		c.Filter.Namespaces = append(c.Filter.Namespaces, namespace)
	}
}

// WithPodRegex adds a pod name regex to the log filter
func WithPodRegex(pattern string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err == nil {
				c.Filter.PodNameRegex = regex
			}
		}
	}
}

// WithContainerRegex adds a container name regex to the log filter
func WithContainerRegex(pattern string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err == nil {
				c.Filter.ContainerRegex = regex
			}
		}
	}
}

// WithLabel adds a label selector to the log filter
func WithLabel(key, value string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if key != "" {
			sel := labels.SelectorFromSet(labels.Set{key: value})
			c.Filter.LabelSelector = sel
		}
	}
}

// WithLabelSelector adds a label selector string to the log filter
// The format is the same as kubectl's label selector (e.g., "app=myapp,env=prod")
func WithLabelSelector(selector string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if selector != "" {
			sel, err := labels.Parse(selector)
			if err == nil {
				c.Filter.LabelSelector = sel
			}
		}
	}
}

// WithIncludeRegex adds an include regex to the log filter
func WithIncludeRegex(pattern string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err == nil {
				c.Filter.IncludeRegex = regex
			}
		}
	}
}

// WithSince sets the time to stream logs from
func WithSince(duration time.Duration) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if duration >= 0 {
			tm := time.Now().Add(-duration)
			c.Filter.Since = &tm
		}
	}
}

// WithContainerState sets the container state filter
func WithContainerState(state string) StreamOption {
	return func(c *StreamConfig) {
		if c.Filter == nil {
			c.Filter = &LogFilter{}
		}
		if state != "" {
			c.Filter.ContainerState = state
		}
	}
}
