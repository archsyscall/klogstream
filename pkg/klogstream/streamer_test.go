package klogstream

import (
	"bytes"
	"context"
	"regexp"
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

// MockStreamer is a mock implementation of the Streamer interface for testing
type MockStreamer struct {
	StartCalled bool
	StopCalled  bool
	StartError  error
	mu          sync.Mutex
}

func (m *MockStreamer) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StartCalled = true
	return m.StartError
}

func (m *MockStreamer) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StopCalled = true
}

// MockFactory is used to create mock streamers for testing
type MockFactory struct {
	CreateFunc func(options ...StreamOption) (Streamer, error)
}

func (f *MockFactory) NewStreamer(options ...StreamOption) (Streamer, error) {
	return f.CreateFunc(options...)
}

func TestNewStreamer_NeedsFilter(t *testing.T) {
	handler := NewConsoleHandler()
	restConfig := &rest.Config{
		Host: "https://test-server:8443",
	}

	_, err := NewStreamer(
		WithRestConfig(restConfig),
		WithHandler(handler),
	)

	if err == nil {
		t.Error("Expected error for missing filter, got none")
	}
}

func TestNewStreamer_NeedsHandler(t *testing.T) {
	filter, err := NewLogFilterBuilder().
		Namespace("default").
		Build()
	if err != nil {
		t.Fatal(err)
	}

	restConfig := &rest.Config{
		Host: "https://test-server:8443",
	}

	_, err = NewStreamer(
		WithRestConfig(restConfig),
		WithFilter(filter),
	)

	if err == nil {
		t.Error("Expected error for missing handler, got none")
	}
}

func TestStreamBuilder(t *testing.T) {
	origNewStreamer := NewStreamer
	defer func() {
		NewStreamer = origNewStreamer
	}()

	mockStreamer := &MockStreamer{}
	mockFactory := &MockFactory{
		CreateFunc: func(options ...StreamOption) (Streamer, error) {
			return mockStreamer, nil
		},
	}

	NewStreamer = mockFactory.NewStreamer

	// Test the builder
	builder := NewBuilder().
		WithNamespace("default").
		WithPodRegex("nginx-").
		WithContainerRegex("web").
		WithLabel("app", "web").
		WithIncludeRegex("ERROR").
		WithFormatter(NewTextFormatter()).
		WithHandler(NewConsoleHandler()).
		WithMatcher(NewJavaStackMatcher())

	// Build and verify
	streamer, err := builder.Build()
	if err != nil {
		t.Fatalf("Builder.Build() error = %v", err)
	}

	// Start and verify
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = streamer.Start(ctx)
	if err != nil {
		t.Fatalf("Streamer.Start() error = %v", err)
	}

	if !mockStreamer.StartCalled {
		t.Error("Streamer.Start() not called")
	}

	// Stop and verify
	streamer.Stop()
	if !mockStreamer.StopCalled {
		t.Error("Streamer.Stop() not called")
	}
}

func TestRun(t *testing.T) {
	origNewStreamer := NewStreamer
	defer func() {
		NewStreamer = origNewStreamer
	}()

	mockStreamer := &MockStreamer{}
	mockFactory := &MockFactory{
		CreateFunc: func(options ...StreamOption) (Streamer, error) {
			return mockStreamer, nil
		},
	}

	NewStreamer = mockFactory.NewStreamer

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Run with some options
	err := Run(ctx,
		WithNamespace("default"),
		WithPodRegex("nginx-"),
		WithHandler(NewConsoleHandler()),
	)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !mockStreamer.StartCalled {
		t.Error("Run(): Streamer.Start() not called")
	}

	if !mockStreamer.StopCalled {
		t.Error("Run(): Streamer.Stop() not called")
	}
}

func TestBuilderRun(t *testing.T) {
	origNewStreamer := NewStreamer
	defer func() {
		NewStreamer = origNewStreamer
	}()

	mockStreamer := &MockStreamer{}
	mockFactory := &MockFactory{
		CreateFunc: func(options ...StreamOption) (Streamer, error) {
			return mockStreamer, nil
		},
	}

	NewStreamer = mockFactory.NewStreamer

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Use the builder's Run method
	err := NewBuilder().
		WithNamespace("default").
		WithPodRegex("nginx-").
		WithHandler(NewConsoleHandler()).
		Run(ctx)

	if err != nil {
		t.Fatalf("Builder.Run() error = %v", err)
	}

	if !mockStreamer.StartCalled {
		t.Error("Builder.Run(): Streamer.Start() not called")
	}

	if !mockStreamer.StopCalled {
		t.Error("Builder.Run(): Streamer.Stop() not called")
	}
}

func TestOptionsBuilder(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*StreamConfig)
		verifyFunc func(*testing.T, *StreamConfig)
	}{
		{
			name: "WithNamespace",
			setupFunc: func(c *StreamConfig) {
				option := WithNamespace("test-namespace")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if len(c.Filter.Namespaces) != 1 || c.Filter.Namespaces[0] != "test-namespace" {
					t.Errorf("WithNamespace() did not set namespace correctly, got %v", c.Filter.Namespaces)
				}
			},
		},
		{
			name: "WithPodRegex",
			setupFunc: func(c *StreamConfig) {
				option := WithPodRegex("pod-.*")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.PodNameRegex == nil || c.Filter.PodNameRegex.String() != "pod-.*" {
					t.Errorf("WithPodRegex() did not set pod regex correctly, got %v",
						c.Filter.PodNameRegex)
				}
			},
		},
		{
			name: "WithContainerRegex",
			setupFunc: func(c *StreamConfig) {
				option := WithContainerRegex("container-.*")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.ContainerRegex == nil || c.Filter.ContainerRegex.String() != "container-.*" {
					t.Errorf("WithContainerRegex() did not set container regex correctly, got %v",
						c.Filter.ContainerRegex)
				}
			},
		},
		{
			name: "WithLabel",
			setupFunc: func(c *StreamConfig) {
				option := WithLabel("app", "web")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.LabelSelector == nil {
					t.Errorf("WithLabel() did not set label selector")
				} else {
					requirements, _ := c.Filter.LabelSelector.Requirements()
					if len(requirements) != 1 || requirements[0].Key() != "app" ||
						requirements[0].Values().Has("web") != true {
						t.Errorf("WithLabel() did not set label selector correctly, got %v",
							c.Filter.LabelSelector)
					}
				}
			},
		},
		{
			name: "WithIncludeRegex",
			setupFunc: func(c *StreamConfig) {
				option := WithIncludeRegex("ERROR")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.IncludeRegex == nil || c.Filter.IncludeRegex.String() != "ERROR" {
					t.Errorf("WithIncludeRegex() did not set include regex correctly, got %v",
						c.Filter.IncludeRegex)
				}
			},
		},
		{
			name: "WithSince",
			setupFunc: func(c *StreamConfig) {
				option := WithSince(1 * time.Hour)
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.Since == nil {
					t.Errorf("WithSince() did not set since time")
				} else {
					// Roughly an hour ago (within a minute)
					now := time.Now()
					diff := now.Sub(*c.Filter.Since)
					if diff < 59*time.Minute || diff > 61*time.Minute {
						t.Errorf("WithSince() did not set since time correctly, got %v", *c.Filter.Since)
					}
				}
			},
		},
		{
			name: "WithContainerState",
			setupFunc: func(c *StreamConfig) {
				option := WithContainerState("running")
				option(c)
			},
			verifyFunc: func(t *testing.T, c *StreamConfig) {
				if c.Filter.ContainerState != "running" {
					t.Errorf("WithContainerState() did not set container state correctly, got %v",
						c.Filter.ContainerState)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &StreamConfig{
				Filter: &LogFilter{},
			}
			tt.setupFunc(config)
			tt.verifyFunc(t, config)
		})
	}
}

// TestClientCreation tests that basic client creation works
// This is a very basic test that doesn't connect to a kubernetes cluster
func TestClientCreation(t *testing.T) {
	// We can't easily test actual Kubernetes client creation without a real cluster,
	// so we'll just test that the option builder does its job

	// Create a fake rest config
	restConfig := &rest.Config{
		Host: "https://test-server:8443",
	}

	// Create a test filter with namespace (required)
	filter := &LogFilter{
		PodNameRegex: regexp.MustCompile("nginx-.*"),
		Namespaces:   []string{"default"},
	}

	// Create a test handler
	var buf bytes.Buffer
	handler := NewConsoleHandlerWithWriters(&buf, &buf)

	// Create a streamer config manually
	config := &StreamConfig{}
	WithRestConfig(restConfig)(config)
	WithFilter(filter)(config)
	WithHandler(handler)(config)

	// Verify config
	if config.Filter != filter {
		t.Errorf("WithFilter() did not set filter correctly")
	}

	if config.Handler != handler {
		t.Errorf("WithHandler() did not set handler correctly")
	}

	if len(config.KubeOptions) == 0 {
		t.Errorf("WithRestConfig() did not add kube option")
	}
}
