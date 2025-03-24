package formatter

import (
	"encoding/json"
	"time"
)

// JSONFormatter formats log messages as JSON
type JSONFormatter struct {
	// Fields to include in the JSON output
	IncludeTimestamp     bool
	IncludeNamespace     bool
	IncludePodName       bool
	IncludeContainerName bool
}

// JSONLogEntry represents a log entry in JSON format
type JSONLogEntry struct {
	Timestamp     string `json:"timestamp,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	PodName       string `json:"pod_name,omitempty"`
	ContainerName string `json:"container_name,omitempty"`
	Message       string `json:"message"`
}

// NewJSONFormatter creates a new JSONFormatter with default settings
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		IncludeTimestamp:     true,
		IncludeNamespace:     true,
		IncludePodName:       true,
		IncludeContainerName: true,
	}
}

// Format converts a LogMessage to a JSON string
func (f *JSONFormatter) Format(msg LogMessage) string {
	entry := JSONLogEntry{
		Message: msg.Message,
	}

	if f.IncludeTimestamp {
		entry.Timestamp = msg.Timestamp.Format(time.RFC3339)
	}

	if f.IncludeNamespace {
		entry.Namespace = msg.Namespace
	}

	if f.IncludePodName {
		entry.PodName = msg.PodName
	}

	if f.IncludeContainerName {
		entry.ContainerName = msg.ContainerName
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback in case of marshaling error
		return msg.Message
	}

	return string(data)
}
