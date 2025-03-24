# klogstream Tests

This directory contains integration tests for the klogstream library.

## Test Structure

### Integration Tests

The integration tests use the Kubernetes Go client's testing utilities to verify the functionality of klogstream without requiring an actual Kubernetes cluster. These tests ensure that:

1. Proper filter construction and behavior
2. API compatibility with Kubernetes types
3. Core functionality works as expected

#### Using fake.Clientset

For testing with a simulated Kubernetes environment, the `fake.Clientset` from the Kubernetes client-go library can be used:

```go
// Create a fake clientset
clientset := fake.NewSimpleClientset()

// Create test resources
clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})

// Use with klogstream
streamer, err := klogstream.NewBuilder().
    WithClientset(clientset).  // Directly inject the fake clientset
    WithNamespace(namespace).
    WithPodRegex("test-.*").
    WithHandler(handler).
    Build()
```

## Running Tests

To run all tests, including unit tests and integration tests:

```bash
go test -v ./...
```

To run only the integration tests:

```bash
go test -v ./test/integration
```

## Testing with a Real Cluster

For manual testing with a real Kubernetes cluster, see the examples in the `examples/` directory, especially the `examples/k3d/` directory which includes instructions for setting up a test cluster using k3d and deploying test applications.

The k3d examples are particularly useful for:
1. Testing log streaming from real Kubernetes pods
2. Validating multiline log handling
3. Testing custom formatters and handlers in a real environment