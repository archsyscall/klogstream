package klogstream

import (
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

func TestStreamOptions(t *testing.T) {
	// Create a test config
	restConfig := &rest.Config{
		Host: "https://test-server:8443",
	}

	// Create a test filter
	filter := &LogFilter{
		Namespaces: []string{"default"},
	}

	// Create a test formatter
	formatter := NewTextFormatter()

	// Create a test handler
	handler := NewConsoleHandler()

	// Create a test matcher
	matcher := NewJavaStackMatcher()

	// Create a test retry policy
	retryPolicy := RetryPolicy{
		MaxRetries:      10,
		InitialInterval: 2 * time.Second,
		MaxInterval:     60 * time.Second,
		Multiplier:      1.5,
	}

	// Create a new stream config with options
	config := NewStreamConfig()

	// Apply all options
	options := []StreamOption{
		WithRestConfig(restConfig),
		WithKubeContext("test-context"),
		WithKubeconfigPath("/path/to/kubeconfig"),
		WithFilter(filter),
		WithFormatter(formatter),
		WithHandler(handler),
		WithMatcher(matcher),
		WithRetryPolicy(retryPolicy),
	}

	// Apply options
	for _, option := range options {
		option(config)
	}

	// Verify all options were applied
	if config.Filter != filter {
		t.Errorf("WithFilter option was not applied correctly")
	}

	if config.Formatter != formatter {
		t.Errorf("WithFormatter option was not applied correctly")
	}

	if config.Handler != handler {
		t.Errorf("WithHandler option was not applied correctly")
	}

	if config.Matcher != matcher {
		t.Errorf("WithMatcher option was not applied correctly")
	}

	if config.RetryPolicy.MaxRetries != retryPolicy.MaxRetries {
		t.Errorf("WithRetryPolicy option was not applied correctly")
	}

	// The KubeOptions should have been appended
	if len(config.KubeOptions) != 4 { // Default option + 3 added
		t.Errorf("Expected 4 KubeOptions, got %d", len(config.KubeOptions))
	}
}
