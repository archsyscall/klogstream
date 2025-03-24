package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/archsyscall/klogstream/pkg/klogstream"
)

// DebugHandler prints detailed information for debugging
type DebugHandler struct{}

func (h *DebugHandler) OnLog(message klogstream.LogMessage) {
	fmt.Printf("[LOG] Pod: %s, Container: %s, Timestamp: %s, Message: %s\n",
		message.PodName,
		message.ContainerName,
		message.Timestamp.Format(time.RFC3339),
		message.Message)
}

func (h *DebugHandler) OnError(err error) {
	fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
}

func (h *DebugHandler) OnEnd() {
	fmt.Println("[END] Streaming ended")
}

func main() {
	// Create a context with timeout to avoid waiting indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Print detailed info about the kubernetes environment
	fmt.Println("Starting debug log streamer...")
	fmt.Println("Kubernetes context information:")
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}
	fmt.Printf("- KUBECONFIG: %s\n", kubeconfigPath)

	// List all pods to see what's available
	fmt.Println("\nAvailable pods (via kubectl):")
	cmd := exec.Command("kubectl", "get", "pods", "-A")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running kubectl: %v\n", err)
	} else {
		fmt.Println(string(output))
	}

	// Create a streamer with detailed debug output
	builder := klogstream.NewBuilder()

	// Try multiple namespace options to make sure we capture something
	builder.WithNamespace("default")

	// Set debug handler
	builder.WithHandler(&DebugHandler{})

	// Build the streamer
	streamer, err := builder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating streamer: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nStarting to stream logs...")
	fmt.Println("Press Ctrl+C to stop")

	// Start streaming in a goroutine
	errCh := make(chan error, 1)
	go func() {
		fmt.Println("Starting stream in goroutine...")
		err := streamer.Start(ctx)
		fmt.Printf("Stream completed with error: %v\n", err)
		errCh <- err
	}()

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Error streaming logs: %v\n", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		fmt.Println("Context deadline exceeded or cancelled")
	}

	fmt.Println("Debug session completed")
}
