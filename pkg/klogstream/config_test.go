package klogstream

import (
	"regexp"
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

// MockHandler implements LogHandler for testing
type MockHandler struct{}

func (h *MockHandler) OnLog(msg LogMessage) {}
func (h *MockHandler) OnError(err error)    {}
func (h *MockHandler) OnEnd()               {}

// MockFormatter implements LogFormatter for testing
type MockFormatter struct{}

func (f *MockFormatter) Format(msg LogMessage) string { return msg.Message }

// MockMatcher implements MultilineMatcher for testing
type MockMatcher struct{}

func (m *MockMatcher) ShouldMerge(previous, next string) bool { return false }

func TestConfigBuilder_Build(t *testing.T) {
	// Create a test filter
	filter := &LogFilter{
		PodNameRegex:   regexp.MustCompile("test"),
		ContainerState: "all",
		Namespaces:     []string{"default"},
	}

	// Create a test kube config
	kubeConfig := &rest.Config{
		Host: "https://localhost:8443",
	}

	tests := []struct {
		name      string
		buildFunc func(*ConfigBuilder) *ConfigBuilder
		wantErr   bool
	}{
		{
			name: "empty builder",
			buildFunc: func(b *ConfigBuilder) *ConfigBuilder {
				return b
			},
			wantErr: true,
		},
		{
			name: "missing filter",
			buildFunc: func(b *ConfigBuilder) *ConfigBuilder {
				return b.WithKubeConfig(kubeConfig).
					WithHandler(&MockHandler{})
			},
			wantErr: true,
		},
		{
			name: "missing handler",
			buildFunc: func(b *ConfigBuilder) *ConfigBuilder {
				return b.WithKubeConfig(kubeConfig).
					WithFilter(filter)
			},
			wantErr: true,
		},
		{
			name: "missing kube config",
			buildFunc: func(b *ConfigBuilder) *ConfigBuilder {
				return b.WithFilter(filter).
					WithHandler(&MockHandler{})
			},
			wantErr: true,
		},
		{
			name: "complete config",
			buildFunc: func(b *ConfigBuilder) *ConfigBuilder {
				return b.WithKubeConfig(kubeConfig).
					WithFilter(filter).
					WithHandler(&MockHandler{}).
					WithFormatter(&MockFormatter{}).
					WithMatcher(&MockMatcher{}).
					WithRetryPolicy(RetryPolicy{
						MaxRetries:      10,
						InitialInterval: 2 * time.Second,
						MaxInterval:     60 * time.Second,
						Multiplier:      1.5,
					})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewConfigBuilder()
			builder = tt.buildFunc(builder)
			config, err := builder.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigBuilder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Check fields in complete config case
				if config.KubeConfig == nil {
					t.Errorf("KubeConfig is nil")
				}
				if config.Filter == nil {
					t.Errorf("Filter is nil")
				}
				if config.Handler == nil {
					t.Errorf("Handler is nil")
				}
			}
		})
	}
}

func TestConfigBuilder_Chaining(t *testing.T) {
	// Create a test filter
	filter := &LogFilter{
		PodNameRegex:   regexp.MustCompile("test"),
		ContainerState: "all",
		Namespaces:     []string{"default"},
	}

	// Create a test kube config
	kubeConfig := &rest.Config{
		Host: "https://localhost:8443",
	}

	// Create a custom retry policy
	customRetry := RetryPolicy{
		MaxRetries:      10,
		InitialInterval: 2 * time.Second,
		MaxInterval:     60 * time.Second,
		Multiplier:      1.5,
	}

	// Test chaining methods
	builder := NewConfigBuilder()
	config, err := builder.
		WithKubeConfig(kubeConfig).
		WithFilter(filter).
		WithHandler(&MockHandler{}).
		WithFormatter(&MockFormatter{}).
		WithMatcher(&MockMatcher{}).
		WithRetryPolicy(customRetry).
		Build()

	// Verify no error
	if err != nil {
		t.Fatalf("ConfigBuilder.Build() unexpected error: %v", err)
	}

	// Verify config values
	if config.KubeConfig != kubeConfig {
		t.Errorf("KubeConfig not set correctly")
	}

	if config.Filter != filter {
		t.Errorf("Filter not set correctly")
	}

	if _, ok := config.Handler.(*MockHandler); !ok {
		t.Errorf("Handler not set correctly")
	}

	if _, ok := config.Formatter.(*MockFormatter); !ok {
		t.Errorf("Formatter not set correctly")
	}

	if _, ok := config.Matcher.(*MockMatcher); !ok {
		t.Errorf("Matcher not set correctly")
	}

	if config.RetryPolicy.MaxRetries != customRetry.MaxRetries {
		t.Errorf("RetryPolicy.MaxRetries not set correctly, got %d, want %d",
			config.RetryPolicy.MaxRetries, customRetry.MaxRetries)
	}

	if config.RetryPolicy.InitialInterval != customRetry.InitialInterval {
		t.Errorf("RetryPolicy.InitialInterval not set correctly")
	}
}
