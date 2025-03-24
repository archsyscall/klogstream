package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/archsyscall/klogstream/pkg/klogstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// JavaStackMatcher is an example custom matcher for Java stack traces
type JavaStackMatcher struct{}

// ShouldMerge checks if a line is part of a multiline log entry
func (m *JavaStackMatcher) ShouldMerge(previous, next string) bool {
	return strings.Contains(previous, "Exception") ||
		strings.HasPrefix(strings.TrimSpace(next), "at ") ||
		strings.HasPrefix(strings.TrimSpace(next), "Caused by:")
}

// JSONLogHandler is an example of a handler that processes JSON logs
type JSONLogHandler struct {
	next klogstream.LogHandler
}

// NewJSONLogHandler creates a new JSON log handler
func NewJSONLogHandler(next klogstream.LogHandler) *JSONLogHandler {
	return &JSONLogHandler{
		next: next,
	}
}

// OnLog processes JSON logs, extracts fields, and passes them to the next handler
func (h *JSONLogHandler) OnLog(message klogstream.LogMessage) {
	// Check if this is a JSON log message
	if !strings.HasPrefix(strings.TrimSpace(message.Message), "{") {
		// Not JSON, pass through to next handler
		h.next.OnLog(message)
		return
	}

	// Parse the JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(message.Message), &data); err != nil {
		// Not valid JSON, pass through
		h.next.OnLog(message)
		return
	}

	// Extract and enhance the message
	level := ""
	if l, ok := data["level"]; ok {
		level = fmt.Sprintf("[%s] ", l)
	}

	messageText := ""
	if msg, ok := data["message"]; ok {
		messageText = fmt.Sprintf("%v", msg)
	}

	traceID := ""
	if id, ok := data["trace_id"]; ok {
		traceID = fmt.Sprintf(" (trace: %v)", id)
	}

	// Create enhanced content
	enhancedContent := fmt.Sprintf("%s%s%s", level, messageText, traceID)

	// For errors, include error details
	if level == "[ERROR] " {
		if errorDetails, ok := data["error_details"]; ok {
			enhancedContent += fmt.Sprintf("\nError details: %v", errorDetails)
		}
	}

	// Create a modified log message
	enhancedMessage := klogstream.LogMessage{
		PodName:       message.PodName,
		ContainerName: message.ContainerName,
		Namespace:     message.Namespace,
		Message:       enhancedContent,
		Timestamp:     message.Timestamp,
		Raw:           message.Raw,
	}

	// Pass the enhanced message to the next handler
	h.next.OnLog(enhancedMessage)
}

// OnError passes errors to the next handler
func (h *JSONLogHandler) OnError(err error) {
	h.next.OnError(err)
}

// OnEnd passes the end signal to the next handler
func (h *JSONLogHandler) OnEnd() {
	h.next.OnEnd()
}

// ConsoleHandler is a simple handler that prints logs to the console
type ConsoleHandler struct{}

func (h *ConsoleHandler) OnLog(message klogstream.LogMessage) {
	fmt.Printf("[%s] %s/%s: %s\n",
		message.Timestamp.Format(time.RFC3339),
		message.PodName,
		message.ContainerName,
		message.Message)
}

func (h *ConsoleHandler) OnError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func (h *ConsoleHandler) OnEnd() {
	fmt.Println("Log streaming ended")
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

	// Get the Kubernetes clientset
	clientset, err := getClientset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Create options for different deployments
	webAppBuilder := klogstream.NewBuilder().
		WithClientset(clientset).
		WithNamespace("klogstream-demo").
		WithPodLabelSelector("app=web-app").
		WithHandler(&ConsoleHandler{})

	javaAppBuilder := klogstream.NewBuilder().
		WithClientset(clientset).
		WithNamespace("klogstream-demo").
		WithPodLabelSelector("app=java-app").
		WithMatcher(&JavaStackMatcher{}).
		WithHandler(&ConsoleHandler{})

	jsonLoggerBuilder := klogstream.NewBuilder().
		WithClientset(clientset).
		WithNamespace("klogstream-demo").
		WithPodLabelSelector("app=json-logger").
		WithHandler(NewJSONLogHandler(&ConsoleHandler{}))

	// Create three streamers for different app types
	webStreamer, err := webAppBuilder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating web app streamer: %v\n", err)
		os.Exit(1)
	}

	javaStreamer, err := javaAppBuilder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Java app streamer: %v\n", err)
		os.Exit(1)
	}

	jsonStreamer, err := jsonLoggerBuilder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON logger streamer: %v\n", err)
		os.Exit(1)
	}

	// Start streaming in three goroutines
	fmt.Println("Starting to stream logs from the klogstream-demo namespace")
	fmt.Println("Press Ctrl+C to stop")

	errChan := make(chan error, 3)

	go func() {
		errChan <- webStreamer.Start(ctx)
	}()

	go func() {
		errChan <- javaStreamer.Start(ctx)
	}()

	go func() {
		errChan <- jsonStreamer.Start(ctx)
	}()

	// Wait for any streamer to exit or context cancellation
	fmt.Println("Streaming logs... (waiting for log entries)")
	for {
		select {
		case err := <-errChan:
			if err != nil && err != context.Canceled {
				fmt.Fprintf(os.Stderr, "Error streaming logs: %v\n", err)
				os.Exit(1)
			}
		case <-ctx.Done():
			// Expected cancellation
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("Log streaming completed")
}

// getClientset creates a kubernetes clientset from the default kubeconfig
func getClientset() (*kubernetes.Clientset, error) {
	// Get kubeconfig from default location
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home := os.Getenv("HOME")
		kubeconfigPath = fmt.Sprintf("%s/.kube/config", home)
	}

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}
