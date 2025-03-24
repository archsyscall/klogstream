# Custom klogstream Example

This example demonstrates advanced usage of the klogstream library with custom log handlers, formatters, and complex filtering. It streams logs from pods in the kube-system namespace with names starting with "kube-".

## Features

This example demonstrates:

1. **Custom Log Handler**: Writes logs to a file with custom formatting
2. **Custom JSON Formatter**: Formats logs as JSON objects with additional metadata
3. **Multi-Output Handler**: Sends logs to multiple destinations (console and file)
4. **Template-based Formatting**: Uses Go templates to format log output
5. **Advanced Filtering**: Filters logs by namespace, pod name regex, and limits output

## Prerequisites

- A running Kubernetes cluster (k3d, minikube, or any other Kubernetes cluster)
- kubectl configured to access the cluster

## Running the Example

```bash
go run main.go
```

## Expected Output

You should see logs from pods in the kube-system namespace displayed in the console according to the template format. Additionally, logs are written to a file named `kube-system-logs.txt` with sequential numbering.

Press Ctrl+C to stop the log streaming.

## Code Explanation

1. **Custom Log Handler**: `CustomLogHandler` implements the `LogHandler` interface to write logs to a file
2. **Custom JSON Formatter**: `CustomJSONFormatter` implements the `LogFormatter` interface to format logs as JSON
3. **Multi-Output Handler**: `MultiOutputHandler` sends logs to multiple destinations
4. **Template Handler**: `TemplateHandler` formats logs using Go templates
5. **Configuration**: Sets up filtering to only show logs from specific pods
6. **Options**: Configures custom handlers and formatters