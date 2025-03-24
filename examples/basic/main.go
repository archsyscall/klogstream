package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archsyscall/klogstream/pkg/klogstream"
)

// ConsoleLogHandler is a simple handler that prints logs to the console
type ConsoleLogHandler struct{}

func (h *ConsoleLogHandler) OnLog(message klogstream.LogMessage) {
	fmt.Printf("[%s] %s/%s: %s\n",
		message.Timestamp.Format(time.RFC3339),
		message.PodName,
		message.ContainerName,
		message.Message)
}

func (h *ConsoleLogHandler) OnError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func (h *ConsoleLogHandler) OnEnd() {
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

	// Create a streamer using the builder pattern
	streamer, err := klogstream.NewBuilder().
		WithNamespace("default").          // Stream logs from the default namespace
		WithPodRegex(".*").                // Stream logs from all pods
		WithContainerRegex(".*").          // Stream logs from all containers
		WithHandler(&ConsoleLogHandler{}). // Use our custom console handler
		Build()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating streamer: %v\n", err)
		os.Exit(1)
	}

	// Start streaming logs
	fmt.Println("Starting to stream logs from default namespace...")
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
