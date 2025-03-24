package klogstream

import (
	"github.com/archsyscall/klogstream/internal/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// StreamOption is a function that configures a streamer
type StreamOption func(*StreamConfig)

// StreamConfig holds all the configuration for a streamer
type StreamConfig struct {
	// KubeOptions are the options for the kubernetes client
	KubeOptions []kube.Option
	// Filter is the log filter
	Filter *LogFilter
	// Formatter is the log formatter
	Formatter LogFormatter
	// Handler is the log handler
	Handler LogHandler
	// Matcher is the multiline matcher
	Matcher MultilineMatcher
	// RetryPolicy configures retry behavior
	RetryPolicy RetryPolicy
}

// NewStreamConfig creates a new StreamConfig with default values
func NewStreamConfig() *StreamConfig {
	return &StreamConfig{
		KubeOptions: []kube.Option{kube.UseDefaultConfig()},
		RetryPolicy: DefaultRetryPolicy,
	}
}

// WithRestConfig sets the kubernetes client configuration
func WithRestConfig(config *rest.Config) StreamOption {
	return func(c *StreamConfig) {
		c.KubeOptions = append(c.KubeOptions, kube.WithRestConfig(config))
	}
}

// WithKubeconfigPath sets the path to the kubeconfig file
func WithKubeconfigPath(path string) StreamOption {
	return func(c *StreamConfig) {
		c.KubeOptions = append(c.KubeOptions, kube.WithKubeconfigPath(path))
	}
}

// WithKubeContext sets the kubernetes context to use
func WithKubeContext(name string) StreamOption {
	return func(c *StreamConfig) {
		c.KubeOptions = append(c.KubeOptions, kube.WithContextName(name))
	}
}

// WithClientset sets a direct kubernetes clientset to use
// This is especially useful for testing with fake.Clientset
func WithClientset(clientset *kubernetes.Clientset) StreamOption {
	return func(c *StreamConfig) {
		c.KubeOptions = append(c.KubeOptions, kube.WithClientset(clientset))
	}
}

// WithFilter sets the log filter
func WithFilter(filter *LogFilter) StreamOption {
	return func(c *StreamConfig) {
		c.Filter = filter
	}
}

// WithFormatter sets the log formatter
func WithFormatter(formatter LogFormatter) StreamOption {
	return func(c *StreamConfig) {
		c.Formatter = formatter
	}
}

// WithHandler sets the log handler
func WithHandler(handler LogHandler) StreamOption {
	return func(c *StreamConfig) {
		c.Handler = handler
	}
}

// WithMatcher sets the multiline matcher
func WithMatcher(matcher MultilineMatcher) StreamOption {
	return func(c *StreamConfig) {
		c.Matcher = matcher
	}
}

// WithRetryPolicy sets the retry policy
func WithRetryPolicy(policy RetryPolicy) StreamOption {
	return func(c *StreamConfig) {
		c.RetryPolicy = policy
	}
}
