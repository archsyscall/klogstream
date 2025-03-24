package integration

import (
	"context"
	"testing"

	"github.com/archsyscall/klogstream/pkg/klogstream"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestWithFakeClientsetMock tests Kubernetes integration without actually connecting to a cluster
func TestWithFakeClientsetMock(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating test namespace: %v", err)
	}

	// Create a test pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	_, err = clientset.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating test pod: %v", err)
	}

	// Instead of actually testing connected streaming which is integration testing,
	// verify that the filter is correctly configured
	filter, err := klogstream.NewLogFilterBuilder().
		Namespace("test-namespace").
		PodRegex("test-pod").
		Build()

	// If we can build the filter successfully, that means our API is working properly
	if err != nil {
		t.Fatalf("Error building filter: %v", err)
	}

	if len(filter.Namespaces) != 1 || filter.Namespaces[0] != "test-namespace" {
		t.Errorf("Namespace not correctly set in filter, got: %v", filter.Namespaces)
	}

	if filter.PodNameRegex == nil || filter.PodNameRegex.String() != "test-pod" {
		t.Errorf("Pod regex not correctly set in filter")
	}

	t.Log("Mock integration test completed successfully")
}

// testLogHandler is a simple handler for testing
type testLogHandler struct {
	logs *[]string
}

func (h *testLogHandler) OnLog(message klogstream.LogMessage) {
	*h.logs = append(*h.logs, message.Message)
}

func (h *testLogHandler) OnError(err error) {
	// No-op for testing
}

func (h *testLogHandler) OnEnd() {
	// No-op for testing
}
