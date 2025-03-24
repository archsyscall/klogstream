# klogstream

**klogstream** is a **Go library** that implements **Kubernetes** log streaming with **multi-pod filtering**, **JSON formatting**, and **error handling**. The library extends standard Kubernetes clients by adding **regex-based filtering** and **concurrent log processing**. It provides a handler interface for custom log processing, automatic reconnection mechanisms, and built-in support for multiline log assembly. Developers can filter logs by pod names, namespaces, and labels while streaming from multiple containers simultaneously.

## Features

- Concurrent log streaming across multiple pods/containers using goroutines
- Regex-based filtering for pod/container names
- Namespace and label-based log filtering
- Multiline log reassembly (e.g., Java stack traces)
- Flexible log formatting with JSON and custom formats
- Pluggable log handler system
- Automatic reconnection with exponential backoff
- Direct Kubernetes clientset injection support for testing

## Installation

```bash
go get github.com/archsyscall/klogstream/pkg/klogstream
```

## Log Handler Interface

klogstream provides a LogHandler interface for flexible log processing:

```go
type LogHandler interface {
    // OnLog is called whenever a new log message arrives
    OnLog(message LogMessage)
    
    // OnError is called when an error occurs during streaming
    OnError(err error)
    
    // OnEnd is called when the streaming ends
    OnEnd()
}
```

LogMessage structure:
```go
type LogMessage struct {
    Timestamp     time.Time
    PodName       string
    ContainerName string
    Message       string
    Labels        map[string]string
}
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"
	// Other imports...

	"github.com/archsyscall/klogstream/pkg/klogstream"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Set up context and signal handling...

	// Create a streamer using the builder pattern - this is the main part!
	streamer, err := klogstream.NewBuilder().
		WithNamespace("default").             // Stream logs from the default namespace
		WithPodRegex("my-app.*").             // Stream logs from pods matching regex
		WithContainerRegex(".*").             // Stream logs from all containers
		WithHandler(&ConsoleHandler{}).       // Use a custom log handler
		Build()

	if err != nil {
		// Error handling...
	}

	// Start streaming logs
	if err := streamer.Start(ctx); err != nil && err != context.Canceled {
		// Error handling...
	}
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

// OnError and OnEnd implementations...
```

## Development Status

This project is under active development and not yet production-ready. APIs may change without notice.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT](LICENSE)