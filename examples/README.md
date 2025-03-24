# klogstream Examples

This directory contains examples of how to use the klogstream library for streaming logs from Kubernetes pods.

## Prerequisites

1. **k3d** - A lightweight wrapper to run k3s (Rancher Lab's minimal Kubernetes distribution) in docker
   ```
   # MacOS
   brew install k3d
   
   # Linux
   curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
   ```

2. **kubectl** - The Kubernetes command-line tool
   ```
   # MacOS
   brew install kubectl
   
   # Linux
   curl -LO "https://dl.k8s.io/release/stable.txt"
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x kubectl
   sudo mv kubectl /usr/local/bin/
   ```

## Setting up a test cluster

Create a local k3d cluster for testing:

```bash
k3d cluster create klogstream-test --agents 1
```

## Example Applications

1. [Basic Example](./basic/) - Simple example demonstrating basic usage with the default settings
2. [Custom Example](./custom/) - More complex example showing custom formatters and handlers
3. [k3d Example](./k3d/) - Complete example with k3d setup and demonstration of log streaming

## Running the Examples

Each example directory contains a README with specific instructions.

For the k3d example, follow these steps:

1. Set up the test cluster as described above
2. Deploy the test application:
   ```bash
   kubectl apply -f examples/k3d/manifests/
   ```
3. Run the example:
   ```bash
   go run examples/k3d/main.go
   ```

## Cleaning Up

To delete the test cluster:

```bash
k3d cluster delete klogstream-test
```