package filter

import "errors"

// Error definitions for filter package
var (
	// ErrInvalidRegex is returned when a regex pattern is invalid
	ErrInvalidRegex = errors.New("invalid regular expression pattern")
	// ErrInvalidSinceTime is returned when the since time is invalid
	ErrInvalidSinceTime = errors.New("since time cannot be in the future")
	// ErrInvalidSinceDuration is returned when the since duration is invalid
	ErrInvalidSinceDuration = errors.New("since duration cannot be negative")
	// ErrInvalidContainerState is returned when the container state is invalid
	ErrInvalidContainerState = errors.New("invalid container state, must be 'all', 'running', or 'terminated'")
	// ErrEmptyFilter is returned when no filter criteria are provided
	ErrEmptyFilter = errors.New("at least one filter criteria must be specified")
	// ErrNoNamespaceSpecified is returned when no namespace is specified
	ErrNoNamespaceSpecified = errors.New("no namespace specified")
)
