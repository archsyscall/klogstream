package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClientProvider handles Kubernetes client creation and configuration
type ClientProvider struct {
	// RestConfig is the kubernetes client configuration
	RestConfig *rest.Config
	// KubeconfigPath is the path to the kubeconfig file
	KubeconfigPath string
	// ContextName is the kubernetes context to use
	ContextName string
	// UseInClusterConfig indicates whether to use in-cluster configuration
	UseInClusterConfig bool
	// Clientset is a direct Kubernetes clientset instance
	Clientset *kubernetes.Clientset
}

// NewClientProvider creates a new ClientProvider with default settings
func NewClientProvider() *ClientProvider {
	return &ClientProvider{
		UseInClusterConfig: true,
	}
}

// WithRestConfig sets a specific rest.Config for the client
func (p *ClientProvider) WithRestConfig(config *rest.Config) *ClientProvider {
	p.RestConfig = config
	p.UseInClusterConfig = false
	p.Clientset = nil // Clear any direct clientset when setting config
	return p
}

// WithClientset sets a direct kubernetes clientset
func (p *ClientProvider) WithClientset(clientset *kubernetes.Clientset) *ClientProvider {
	p.Clientset = clientset
	p.UseInClusterConfig = false
	return p
}

// WithKubeconfigPath sets the path to the kubeconfig file
func (p *ClientProvider) WithKubeconfigPath(path string) *ClientProvider {
	p.KubeconfigPath = path
	p.UseInClusterConfig = false
	return p
}

// WithContextName sets the kubernetes context to use
func (p *ClientProvider) WithContextName(name string) *ClientProvider {
	p.ContextName = name
	p.UseInClusterConfig = false
	return p
}

// getDefaultKubeconfigPath returns the default path to the kubeconfig file
func getDefaultKubeconfigPath() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// GetConfig returns a kubernetes rest.Config based on the provider settings
func (p *ClientProvider) GetConfig() (*rest.Config, error) {
	// Case 1: Use provided RestConfig if available
	if p.RestConfig != nil {
		return p.RestConfig, nil
	}

	// Case 2: Try in-cluster config if enabled
	if p.UseInClusterConfig {
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, nil
		}
		// Fall through to kubeconfig if in-cluster config fails
	}

	// Case 3: Use kubeconfig path if provided, or default path
	kubeconfigPath := p.KubeconfigPath
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
		if kubeconfigPath == "" {
			return nil, fmt.Errorf("unable to locate kubeconfig")
		}
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}

	// Load config with or without specific context
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	if p.ContextName != "" {
		configOverrides.CurrentContext = p.ContextName
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Get the config
	return clientConfig.ClientConfig()
}

// GetClientset returns a kubernetes clientset based on the provider settings
func (p *ClientProvider) GetClientset() (*kubernetes.Clientset, error) {
	// If a direct clientset is provided, use it
	if p.Clientset != nil {
		return p.Clientset, nil
	}

	// Otherwise, create a clientset from the config
	config, err := p.GetConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetCurrentNamespace returns the current namespace from the context
func (p *ClientProvider) GetCurrentNamespace() (string, error) {
	// If we're using in-cluster config, use "default" namespace
	if p.UseInClusterConfig && p.RestConfig != nil {
		return "default", nil
	}

	// Use kubeconfig path if provided, or default path
	kubeconfigPath := p.KubeconfigPath
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
		if kubeconfigPath == "" {
			return "", fmt.Errorf("unable to locate kubeconfig")
		}
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return "", fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}

	// Load config with or without specific context
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	if p.ContextName != "" {
		configOverrides.CurrentContext = p.ContextName
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Get namespace from the client config
	namespace, _, err := clientConfig.Namespace()
	return namespace, err
}
