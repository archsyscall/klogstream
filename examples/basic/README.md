# Basic klogstream Example

This example demonstrates the basic usage of the klogstream library with default settings. It streams logs from all pods in the default namespace.

## Prerequisites

- A running Kubernetes cluster (k3d, minikube, or any other Kubernetes cluster)
- kubectl configured to access the cluster

## Running the Example

```bash
# Make sure you have some pods running in the default namespace
# If not, you can deploy a simple nginx pod:
kubectl run nginx --image=nginx

# Run the example
go run main.go
```

## Expected Output

You should see logs from all pods in the default namespace with timestamps, pod names, and container names included. The output will be colorized for better readability.

Press Ctrl+C to stop the log streaming.

## Code Explanation

1. **Setting up a context with cancellation**: Allows for graceful shutdown
2. **Signal handling**: Captures Ctrl+C to stop streaming
3. **Configuration**: Specifies what logs to stream (namespace, pod and container selection)
4. **Options**: Configures output formatting (colors, timestamps, etc.)
5. **Streaming**: Starts the log streaming process