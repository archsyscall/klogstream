package formatter

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONFormatter_Format(t *testing.T) {
	// Create a fixed timestamp for testing
	fixedTime := time.Date(2023, 4, 15, 12, 34, 56, 0, time.UTC)

	// Create a sample message
	msg := LogMessage{
		Namespace:     "default",
		PodName:       "test-pod",
		ContainerName: "test-container",
		Timestamp:     fixedTime,
		Message:       "Test message",
	}

	tests := []struct {
		name                 string
		includeTimestamp     bool
		includeNamespace     bool
		includePodName       bool
		includeContainerName bool
		checkFields          map[string]interface{}
	}{
		{
			name:                 "full json",
			includeTimestamp:     true,
			includeNamespace:     true,
			includePodName:       true,
			includeContainerName: true,
			checkFields: map[string]interface{}{
				"timestamp":      "2023-04-15T12:34:56Z",
				"namespace":      "default",
				"pod_name":       "test-pod",
				"container_name": "test-container",
				"message":        "Test message",
			},
		},
		{
			name:                 "message only",
			includeTimestamp:     false,
			includeNamespace:     false,
			includePodName:       false,
			includeContainerName: false,
			checkFields: map[string]interface{}{
				"message": "Test message",
			},
		},
		{
			name:                 "selective fields",
			includeTimestamp:     true,
			includeNamespace:     false,
			includePodName:       true,
			includeContainerName: false,
			checkFields: map[string]interface{}{
				"timestamp": "2023-04-15T12:34:56Z",
				"pod_name":  "test-pod",
				"message":   "Test message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &JSONFormatter{
				IncludeTimestamp:     tt.includeTimestamp,
				IncludeNamespace:     tt.includeNamespace,
				IncludePodName:       tt.includePodName,
				IncludeContainerName: tt.includeContainerName,
			}

			got := formatter.Format(msg)

			// Parse the JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Check fields
			for k, v := range tt.checkFields {
				if parsed[k] != v {
					t.Errorf("Field %q = %v, want %v", k, parsed[k], v)
				}
			}

			// Check that no extra fields are present
			for k := range parsed {
				if _, ok := tt.checkFields[k]; !ok {
					t.Errorf("Unexpected field %q in JSON", k)
				}
			}
		})
	}
}
