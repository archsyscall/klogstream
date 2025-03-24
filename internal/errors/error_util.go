package errors

import (
	"io"
	"strings"
)

// IsPodDeletedError checks if an error is due to pod deletion
// which should not be treated as an error condition
func IsPodDeletedError(err error) bool {
	if err == nil {
		return false
	}

	// Common error messages when pod is deleted
	if err == io.EOF {
		return true
	}

	errStr := err.Error()
	// Check for common error messages when pod is deleted
	if strings.Contains(errStr, "container not found") ||
		strings.Contains(errStr, "pod not found") ||
		strings.Contains(errStr, "has been terminated") ||
		strings.Contains(errStr, "has been deleted") {
		return true
	}

	return false
}

// IsPermError checks if an error should be considered permanent
func IsPermError(err error) bool {
	// TODO: Implement better detection of permanent errors
	return false
}
