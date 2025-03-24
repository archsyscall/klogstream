package integration

import (
	"testing"

	"github.com/archsyscall/klogstream/pkg/klogstream"
)

// TestSimpleIntegrationMock tests integration without actually connecting to a cluster
func TestSimpleIntegrationMock(t *testing.T) {
	// Verify that we can build a filter
	filter, err := klogstream.NewLogFilterBuilder().
		Namespace("default").
		PodRegex("test.*").
		ContainerRegex("container-.*").
		Build()

	if err != nil {
		t.Fatalf("Error building filter: %v", err)
	}

	// Verify filter properties
	if len(filter.Namespaces) != 1 || filter.Namespaces[0] != "default" {
		t.Errorf("Namespace not correctly set in filter, got: %v", filter.Namespaces)
	}

	if filter.PodNameRegex == nil || filter.PodNameRegex.String() != "test.*" {
		t.Errorf("Pod regex not correctly set in filter")
	}

	if filter.ContainerRegex == nil || filter.ContainerRegex.String() != "container-.*" {
		t.Errorf("Container regex not correctly set in filter")
	}

	// We're not going to connect to a real cluster, but we'll verify that our API
	// for building streams is working as expected.
	t.Log("Simple integration mock test passed")
}
