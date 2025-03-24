# k3d Example with klogstream

This example demonstrates using klogstream with a local k3d cluster. It includes setup instructions, sample applications that generate different types of logs, and examples of streaming and handling those logs.

## Setup

### 1. Create a k3d cluster

```bash
k3d cluster create klogstream-test --agents 1
```

### 2. Deploy the sample applications

```bash
kubectl apply -f manifests/
```

This will deploy:
- A simple web application that logs HTTP requests
- A Java application that generates stack traces
- A JSON-logging application

### 3. Run the example

```bash
go run main.go
```

## Using LogFilterBuilder

The example below shows how to use `NewLogFilterBuilder()` to create log filters:

```go
// Create a filter for error logs in the web-app
errorFilter, err := klogstream.NewLogFilterBuilder().
    Label("app", "web-app").
    Namespace("klogstream-demo").
    Include("ERROR|error|Error").
    ContainerState("running").
    Since(1 * time.Hour).
    Build()
if err != nil {
    log.Fatalf("Error creating filter: %v", err)
}

// Create a streamer with the filter
streamer, err := klogstream.NewBuilder().
    WithClientset(clientset).
    WithFilter(errorFilter).
    WithHandler(&ConsoleHandler{}).
    Build()
if err != nil {
    log.Fatalf("Error creating streamer: %v", err)
}

// Start streaming
err = streamer.Start(ctx)
```

## Sample Applications

### Web Application

A simple web server that logs HTTP requests in plain text format.

### Java Application

A Java application that periodically generates stack traces to demonstrate multiline log handling.

### JSON Logger

An application that outputs logs in JSON format to demonstrate structured log handling.

## Cleanup

When you're done, delete the cluster:

```bash
k3d cluster delete klogstream-test
```