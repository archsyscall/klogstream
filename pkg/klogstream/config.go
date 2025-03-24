package klogstream

import (
	"time"

	"k8s.io/client-go/rest"
)

// Config holds all configuration for the log streamer
type Config struct {
	// KubeConfig is the kubernetes configuration for connecting to the cluster
	KubeConfig *rest.Config
	// Filter defines the criteria for filtering logs
	Filter *LogFilter
	// Formatter defines how logs are formatted
	Formatter LogFormatter
	// Handler processes the log messages
	Handler LogHandler
	// Matcher determines if log lines should be treated as multiline
	Matcher MultilineMatcher
	// RetryPolicy configures retry behavior for transient errors
	RetryPolicy RetryPolicy
}

// RetryPolicy configures the retry behavior for transient errors
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialInterval is the initial delay between retries
	InitialInterval time.Duration
	// MaxInterval is the maximum delay between retries
	MaxInterval time.Duration
	// Multiplier is the factor by which the delay increases between retries
	Multiplier float64
}

// DefaultRetryPolicy provides reasonable default values for retries
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries:      5,
	InitialInterval: 1 * time.Second,
	MaxInterval:     30 * time.Second,
	Multiplier:      2,
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		RetryPolicy: DefaultRetryPolicy,
	}
}

// ConfigBuilder provides a fluent API for building Config
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new ConfigBuilder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// WithKubeConfig sets the kubernetes configuration
func (b *ConfigBuilder) WithKubeConfig(config *rest.Config) *ConfigBuilder {
	b.config.KubeConfig = config
	return b
}

// WithFilter sets the log filter
func (b *ConfigBuilder) WithFilter(filter *LogFilter) *ConfigBuilder {
	b.config.Filter = filter
	return b
}

// WithFormatter sets the log formatter
func (b *ConfigBuilder) WithFormatter(formatter LogFormatter) *ConfigBuilder {
	b.config.Formatter = formatter
	return b
}

// WithHandler sets the log handler
func (b *ConfigBuilder) WithHandler(handler LogHandler) *ConfigBuilder {
	b.config.Handler = handler
	return b
}

// WithMatcher sets the multiline matcher
func (b *ConfigBuilder) WithMatcher(matcher MultilineMatcher) *ConfigBuilder {
	b.config.Matcher = matcher
	return b
}

// WithRetryPolicy sets the retry policy
func (b *ConfigBuilder) WithRetryPolicy(policy RetryPolicy) *ConfigBuilder {
	b.config.RetryPolicy = policy
	return b
}

// Build creates and validates the Config
func (b *ConfigBuilder) Build() (*Config, error) {
	// Validate required fields
	if b.config.KubeConfig == nil {
		return nil, ErrNoKubeConfig
	}

	if b.config.Filter == nil {
		return nil, ErrNoFilter
	}

	if b.config.Handler == nil {
		return nil, ErrNoHandler
	}

	return b.config, nil
}
