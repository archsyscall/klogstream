package klogstream

import "errors"

// Error definitions
var (
	// ErrNoKubeConfig is returned when no kubernetes config is provided
	ErrNoKubeConfig = errors.New("no kubernetes configuration provided")
	// ErrNoFilter is returned when no log filter is provided
	ErrNoFilter = errors.New("no log filter provided")
	// ErrNoHandler is returned when no log handler is provided
	ErrNoHandler = errors.New("no log handler provided")
	// ErrNoKubeContext is returned when the specified kubernetes context is not found
	ErrNoKubeContext = errors.New("kubernetes context not found")
	// ErrStreamClosed is returned when attempting to use a closed stream
	ErrStreamClosed = errors.New("log stream has been closed")
	// ErrMultilineTimeout is returned when a multiline log times out
	ErrMultilineTimeout = errors.New("timed out waiting for multiline log")
	// ErrTooManyLines is returned when a multiline log exceeds the maximum lines
	ErrTooManyLines = errors.New("multiline log exceeds maximum number of lines")
)
