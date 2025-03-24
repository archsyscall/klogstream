package handler

import (
	"fmt"
	"io"
	"os"
	"sync"
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

// ConsoleHandler outputs logs to the console
type ConsoleHandler struct {
	out    io.Writer
	errOut io.Writer
	mutex  sync.Mutex
}

// NewConsoleHandler creates a new ConsoleHandler with stdout and stderr as default outputs
func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{
		out:    os.Stdout,
		errOut: os.Stderr,
	}
}

// NewConsoleHandlerWithWriters creates a new ConsoleHandler with custom writers
func NewConsoleHandlerWithWriters(out, errOut io.Writer) *ConsoleHandler {
	return &ConsoleHandler{
		out:    out,
		errOut: errOut,
	}
}

// OnLog writes formatted log messages to the configured output writer
func (h *ConsoleHandler) OnLog(msg LogMessage) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	fmt.Fprintln(h.out, msg.Message)
}

// OnError writes error messages to the error output writer
func (h *ConsoleHandler) OnError(err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	fmt.Fprintf(h.errOut, "Error: %v\n", err)
}

// OnEnd is called when the stream ends, does nothing by default
func (h *ConsoleHandler) OnEnd() {
	// No action needed for console handler
}
