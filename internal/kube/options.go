package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Option is a function that configures a ClientProvider
type Option func(*ClientProvider)

// WithRestConfig creates an option to configure a ClientProvider with a specific rest.Config
func WithRestConfig(config *rest.Config) Option {
	return func(provider *ClientProvider) {
		provider.WithRestConfig(config)
	}
}

// WithClientset creates an option to configure a ClientProvider with a direct kubernetes clientset
func WithClientset(clientset *kubernetes.Clientset) Option {
	return func(provider *ClientProvider) {
		provider.WithClientset(clientset)
	}
}

// WithKubeconfigPath creates an option to configure a ClientProvider with a specific kubeconfig path
func WithKubeconfigPath(path string) Option {
	return func(provider *ClientProvider) {
		provider.WithKubeconfigPath(path)
	}
}

// WithContextName creates an option to configure a ClientProvider with a specific context name
func WithContextName(name string) Option {
	return func(provider *ClientProvider) {
		provider.WithContextName(name)
	}
}

// UseDefaultConfig creates an option to configure a ClientProvider to use default in-cluster or
// kubeconfig configuration
func UseDefaultConfig() Option {
	return func(provider *ClientProvider) {
		provider.UseInClusterConfig = true
	}
}

// NewClientProviderWithOptions creates a new ClientProvider with the given options
func NewClientProviderWithOptions(options ...Option) *ClientProvider {
	provider := NewClientProvider()

	for _, option := range options {
		option(provider)
	}

	return provider
}
