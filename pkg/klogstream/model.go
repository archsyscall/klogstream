package klogstream

import (
	"time"
)

// LogMessage represents a single log entry from a kubernetes pod/container
type LogMessage struct {
	// Namespace is the kubernetes namespace of the pod
	Namespace string
	// PodName is the name of the pod
	PodName string
	// ContainerName is the name of the container within the pod
	ContainerName string
	// Timestamp is the time when the log message was created
	Timestamp time.Time
	// Message is the log content
	Message string
	// Raw contains the original bytes of the log message
	Raw []byte
}

// LogStreamError represents an error that occurred during log streaming
type LogStreamError struct {
	// Err is the underlying error
	Err error
	// Permanent indicates if this error is permanent and cannot be recovered from
	Permanent bool
	// Reason is a human-readable description of why the error occurred
	Reason string
}

// Error implements the error interface
func (e *LogStreamError) Error() string {
	if e.Reason != "" {
		return e.Reason + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *LogStreamError) Unwrap() error {
	return e.Err
}
