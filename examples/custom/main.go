package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/archsyscall/klogstream/pkg/klogstream"
)

// CustomLogHandler is an example of a custom log handler that writes to a file
type CustomLogHandler struct {
	file    *os.File
	mu      sync.Mutex
	counter int
}

// NewCustomLogHandler creates a new file-based log handler
func NewCustomLogHandler(filename string) (*CustomLogHandler, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &CustomLogHandler{
		file:    file,
		counter: 0,
	}, nil
}

// OnLog writes the log message to a file
func (h *CustomLogHandler) OnLog(message klogstream.LogMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.counter++

	// Write a custom formatted line to the file
	fmt.Fprintf(h.file, "[%d] %s - %s/%s: %s\n",
		h.counter,
		message.Timestamp.Format(time.RFC3339),
		message.Namespace,
		message.PodName,
		message.Message)
}

// OnError handles errors
func (h *CustomLogHandler) OnError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(h.file, "ERROR: %v\n", err)
}

// OnEnd is called when streaming ends
func (h *CustomLogHandler) OnEnd() {
	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(h.file, "Log streaming ended\n")
}

// Close closes the file
func (h *CustomLogHandler) Close() error {
	return h.file.Close()
}

// CustomJSONFormatter formats logs as JSON objects
type CustomJSONFormatter struct{}

// Format formats a log message as a JSON object
func (f *CustomJSONFormatter) Format(message klogstream.LogMessage) string {
	type jsonLog struct {
		Timestamp   time.Time `json:"timestamp"`
		Namespace   string    `json:"namespace"`
		Pod         string    `json:"pod"`
		Container   string    `json:"container"`
		Message     string    `json:"message"`
		Level       string    `json:"level,omitempty"`
		ElapsedTime int64     `json:"elapsed_ms,omitempty"`
	}

	// Parse the message to extract log level (simplified example)
	level := "INFO"
	if len(message.Message) > 5 {
		prefix := message.Message[0:5]
		if prefix == "ERROR" || prefix == "WARN:" || prefix == "INFO:" || prefix == "DEBUG" {
			level = prefix
		}
	}

	log := jsonLog{
		Timestamp:   message.Timestamp,
		Namespace:   message.Namespace,
		Pod:         message.PodName,
		Container:   message.ContainerName,
		Message:     message.Message,
		Level:       level,
		ElapsedTime: time.Since(message.Timestamp).Milliseconds(),
	}

	data, err := json.Marshal(log)
	if err != nil {
		return err.Error()
	}

	return string(data)
}

// MultiOutputHandler sends logs to multiple destinations
type MultiOutputHandler struct {
	handlers []klogstream.LogHandler
}

// NewMultiOutputHandler creates a handler that sends logs to multiple destinations
func NewMultiOutputHandler(handlers ...klogstream.LogHandler) *MultiOutputHandler {
	return &MultiOutputHandler{
		handlers: handlers,
	}
}

// OnLog sends the log to all configured handlers
func (h *MultiOutputHandler) OnLog(message klogstream.LogMessage) {
	for _, handler := range h.handlers {
		handler.OnLog(message)
	}
}

// OnError sends the error to all configured handlers
func (h *MultiOutputHandler) OnError(err error) {
	for _, handler := range h.handlers {
		handler.OnError(err)
	}
}

// OnEnd notifies all handlers that streaming has ended
func (h *MultiOutputHandler) OnEnd() {
	for _, handler := range h.handlers {
		handler.OnEnd()
	}
}

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create a custom file handler
	fileHandler, err := NewCustomLogHandler("kube-system-logs.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file handler: %v\n", err)
		os.Exit(1)
	}
	defer fileHandler.Close()

	// Create a template for console output
	tmpl, err := template.New("console").Parse("{{.Timestamp.Format \"15:04:05\"}} [{{.Namespace}}] {{.PodName}}/{{.ContainerName}}: {{.Message}}\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating template: %v\n", err)
		os.Exit(1)
	}

	// Create a console handler with the template
	consoleHandler := &TemplateHandler{
		Writer:   os.Stdout,
		Template: tmpl,
	}

	// Create a multi-output handler that sends logs to both console and file
	multiHandler := NewMultiOutputHandler(consoleHandler, fileHandler)

	// Create a streamer using the builder pattern
	streamer, err := klogstream.NewBuilder().
		WithNamespace("kube-system").          // Stream logs from kube-system namespace
		WithPodRegex("kube-.*").               // Only pods starting with "kube-"
		WithContainerRegex(".*").              // All containers
		WithHandler(multiHandler).             // Use our multi-output handler
		WithFormatter(&CustomJSONFormatter{}). // Use custom JSON formatter
		Build()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating streamer: %v\n", err)
		os.Exit(1)
	}

	// Start streaming logs
	fmt.Println("Starting to stream logs from kube-system namespace...")
	fmt.Println("Logs are being written to kube-system-logs.txt")
	fmt.Println("Press Ctrl+C to stop")

	// Start the streamer in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- streamer.Start(ctx)
	}()

	// Wait for an error or context cancellation
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Error streaming logs: %v\n", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		// Context was cancelled, normal shutdown
	}

	fmt.Println("Log streaming completed")
}

// TemplateHandler is a log handler that formats logs using a Go template
type TemplateHandler struct {
	Template *template.Template
	Writer   io.Writer
	mu       sync.Mutex
}

// OnLog formats the log message using the template and writes it to the configured writer
func (h *TemplateHandler) OnLog(message klogstream.LogMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Template.Execute(h.Writer, message)
}

// OnError handles errors
func (h *TemplateHandler) OnError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(h.Writer, "ERROR: %v\n", err)
}

// OnEnd is called when streaming ends
func (h *TemplateHandler) OnEnd() {
	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(h.Writer, "Log streaming ended\n")
}
