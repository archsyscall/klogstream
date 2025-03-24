package kube

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestClientProvider_WithRestConfig(t *testing.T) {
	// Create a test config
	config := &rest.Config{
		Host: "https://test-server:8443",
	}

	// Create a provider and set the config
	provider := NewClientProvider()
	provider.WithRestConfig(config)

	// Check that the config was set
	if provider.RestConfig != config {
		t.Errorf("RestConfig was not set correctly")
	}

	// Check that in-cluster config was disabled
	if provider.UseInClusterConfig {
		t.Errorf("UseInClusterConfig should be false when RestConfig is provided")
	}

	// Get the config and check it matches
	resultConfig, err := provider.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if resultConfig != config {
		t.Errorf("GetConfig() returned wrong config")
	}
}

func TestClientProvider_WithKubeconfigPath(t *testing.T) {
	// Skip test if we can't write to temp dir
	tempDir := os.TempDir()
	if tempDir == "" {
		t.Skip("Skipping test because temp directory is not available")
	}

	// Create a temporary kubeconfig file
	kubeconfigPath := filepath.Join(tempDir, "test-kubeconfig")
	defer os.Remove(kubeconfigPath)

	// Write a basic kubeconfig
	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server:8443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
    namespace: test-namespace
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to write test kubeconfig: %v", err)
	}

	// Create a provider and set the kubeconfig path
	provider := NewClientProvider()
	provider.WithKubeconfigPath(kubeconfigPath)

	// Check that the path was set
	if provider.KubeconfigPath != kubeconfigPath {
		t.Errorf("KubeconfigPath was not set correctly")
	}

	// Check that in-cluster config was disabled
	if provider.UseInClusterConfig {
		t.Errorf("UseInClusterConfig should be false when KubeconfigPath is provided")
	}

	// Get the namespace and check it matches
	namespace, err := provider.GetCurrentNamespace()
	if err != nil {
		t.Fatalf("GetCurrentNamespace() error = %v", err)
	}

	if namespace != "test-namespace" {
		t.Errorf("GetCurrentNamespace() = %v, want %v", namespace, "test-namespace")
	}
}

func TestClientProvider_WithContextName(t *testing.T) {
	// Skip test if we can't write to temp dir
	tempDir := os.TempDir()
	if tempDir == "" {
		t.Skip("Skipping test because temp directory is not available")
	}

	// Create a temporary kubeconfig file
	kubeconfigPath := filepath.Join(tempDir, "test-kubeconfig-context")
	defer os.Remove(kubeconfigPath)

	// Write a kubeconfig with multiple contexts
	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server-1:8443
  name: cluster-1
- cluster:
    server: https://test-server-2:8443
  name: cluster-2
contexts:
- context:
    cluster: cluster-1
    user: user-1
    namespace: namespace-1
  name: context-1
- context:
    cluster: cluster-2
    user: user-2
    namespace: namespace-2
  name: context-2
current-context: context-1
users:
- name: user-1
  user:
    token: token-1
- name: user-2
  user:
    token: token-2
`
	err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to write test kubeconfig: %v", err)
	}

	// Create a provider with specific context
	provider := NewClientProvider()
	provider.WithKubeconfigPath(kubeconfigPath).WithContextName("context-2")

	// Check that the context was set
	if provider.ContextName != "context-2" {
		t.Errorf("ContextName was not set correctly")
	}

	// Get the namespace and check it matches the second context
	namespace, err := provider.GetCurrentNamespace()
	if err != nil {
		t.Fatalf("GetCurrentNamespace() error = %v", err)
	}

	if namespace != "namespace-2" {
		t.Errorf("GetCurrentNamespace() = %v, want %v", namespace, "namespace-2")
	}
}

func TestClientProvider_DefaultConfig(t *testing.T) {
	// This test only verifies the fallback behavior without actually connecting

	// Create a provider with default settings
	provider := NewClientProvider()

	// Check that in-cluster config is enabled by default
	if !provider.UseInClusterConfig {
		t.Errorf("UseInClusterConfig should be true by default")
	}

	// Attempting to get the config might fail if not running in a cluster
	// and no kubeconfig is found, so we don't test that here
}

func TestNewClientProviderWithOptions(t *testing.T) {
	// Create a test config
	config := &rest.Config{
		Host: "https://test-server:8443",
	}

	// Create a provider with options
	provider := NewClientProviderWithOptions(
		WithRestConfig(config),
		WithContextName("test-context"),
	)

	// Check that the options were applied
	if provider.RestConfig != config {
		t.Errorf("RestConfig was not set correctly")
	}

	if provider.ContextName != "test-context" {
		t.Errorf("ContextName was not set correctly")
	}

	if provider.UseInClusterConfig {
		t.Errorf("UseInClusterConfig should be false when RestConfig is provided")
	}
}

func TestClientProvider_GetConfigErrors(t *testing.T) {
	// Test with non-existent kubeconfig path
	provider := NewClientProvider().WithKubeconfigPath("/nonexistent/path/to/kubeconfig")

	_, err := provider.GetConfig()
	if err == nil {
		t.Errorf("GetConfig() should fail with non-existent kubeconfig path")
	}

	// Test with invalid context name (requires a valid kubeconfig file)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping test because home directory is not available")
	}

	defaultKubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	if _, err := os.Stat(defaultKubeconfigPath); os.IsNotExist(err) {
		t.Skip("Skipping test because default kubeconfig file does not exist")
	}

	provider = NewClientProvider().WithContextName("nonexistent-context")

	_, err = provider.GetConfig()
	// Error might be different depending on the environment, so we just check that it fails
	if err == nil {
		// Check if the context actually exists (to avoid false failures)
		config, err := clientcmd.LoadFromFile(defaultKubeconfigPath)
		if err != nil {
			t.Skip("Skipping test because default kubeconfig file could not be loaded")
		}

		if _, exists := config.Contexts["nonexistent-context"]; exists {
			t.Skip("Skipping test because 'nonexistent-context' actually exists")
		}

		t.Errorf("GetConfig() should fail with non-existent context name")
	}
}
