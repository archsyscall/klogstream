package formatter

import (
	"fmt"
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

// TextFormatter formats log messages as text
type TextFormatter struct {
	// ShowTimestamp controls whether to display the timestamp
	ShowTimestamp bool
	// ShowNamespace controls whether to display the namespace
	ShowNamespace bool
	// ShowPodName controls whether to display the pod name
	ShowPodName bool
	// ShowContainerName controls whether to display the container name
	ShowContainerName bool
	// TimestampFormat defines the format for timestamps
	TimestampFormat string
	// ColorOutput enables colorized output
	ColorOutput bool
}

// ColorMap defines ANSI color codes for colorized output
var ColorMap = map[string]string{
	"reset":       "\033[0m",
	"black":       "\033[30m",
	"red":         "\033[31m",
	"green":       "\033[32m",
	"yellow":      "\033[33m",
	"blue":        "\033[34m",
	"magenta":     "\033[35m",
	"cyan":        "\033[36m",
	"white":       "\033[37m",
	"boldBlack":   "\033[1;30m",
	"boldRed":     "\033[1;31m",
	"boldGreen":   "\033[1;32m",
	"boldYellow":  "\033[1;33m",
	"boldBlue":    "\033[1;34m",
	"boldMagenta": "\033[1;35m",
	"boldCyan":    "\033[1;36m",
	"boldWhite":   "\033[1;37m",
}

// DefaultTimestampFormat is the default format for timestamps
const DefaultTimestampFormat = time.RFC3339

// NewTextFormatter creates a new TextFormatter with default settings
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		ShowTimestamp:     true,
		ShowNamespace:     true,
		ShowPodName:       true,
		ShowContainerName: true,
		TimestampFormat:   DefaultTimestampFormat,
		ColorOutput:       true,
	}
}

// Format converts a LogMessage to a formatted string
func (f *TextFormatter) Format(msg LogMessage) string {
	var prefix string

	if f.ShowTimestamp {
		prefix += fmt.Sprintf("%s ", msg.Timestamp.Format(f.TimestampFormat))
	}

	if f.ShowNamespace {
		prefix += fmt.Sprintf("[%s] ", msg.Namespace)
	}

	if f.ShowPodName {
		prefix += fmt.Sprintf("%s", msg.PodName)
	}

	if f.ShowContainerName {
		prefix += fmt.Sprintf("/%s", msg.ContainerName)
	}

	if prefix != "" {
		if f.ColorOutput {
			// Color the prefix with cyan
			prefix = ColorMap["cyan"] + prefix + ColorMap["reset"]
		}
		prefix += ": "
	}

	return prefix + msg.Message
}
