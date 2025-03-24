package formatter

import (
	"strings"
	"testing"
	"time"
)

func TestTextFormatter_Format(t *testing.T) {
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
		name              string
		showTimestamp     bool
		showNamespace     bool
		showPodName       bool
		showContainerName bool
		colorOutput       bool
		want              string
		wantContains      []string
	}{
		{
			name:              "full format with color",
			showTimestamp:     true,
			showNamespace:     true,
			showPodName:       true,
			showContainerName: true,
			colorOutput:       true,
			wantContains: []string{
				"2023-04-15T12:34:56Z",
				"[default]",
				"test-pod/test-container",
				"Test message",
			},
		},
		{
			name:              "full format without color",
			showTimestamp:     true,
			showNamespace:     true,
			showPodName:       true,
			showContainerName: true,
			colorOutput:       false,
			wantContains: []string{
				"2023-04-15T12:34:56Z",
				"[default]",
				"test-pod/test-container",
				"Test message",
			},
		},
		{
			name:              "timestamp only",
			showTimestamp:     true,
			showNamespace:     false,
			showPodName:       false,
			showContainerName: false,
			colorOutput:       false,
			wantContains: []string{
				"2023-04-15T12:34:56Z",
				"Test message",
			},
		},
		{
			name:              "namespace only",
			showTimestamp:     false,
			showNamespace:     true,
			showPodName:       false,
			showContainerName: false,
			colorOutput:       false,
			wantContains: []string{
				"[default]",
				"Test message",
			},
		},
		{
			name:              "pod and container only",
			showTimestamp:     false,
			showNamespace:     false,
			showPodName:       true,
			showContainerName: true,
			colorOutput:       false,
			wantContains: []string{
				"test-pod/test-container",
				"Test message",
			},
		},
		{
			name:              "message only",
			showTimestamp:     false,
			showNamespace:     false,
			showPodName:       false,
			showContainerName: false,
			colorOutput:       false,
			want:              "Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &TextFormatter{
				ShowTimestamp:     tt.showTimestamp,
				ShowNamespace:     tt.showNamespace,
				ShowPodName:       tt.showPodName,
				ShowContainerName: tt.showContainerName,
				TimestampFormat:   time.RFC3339,
				ColorOutput:       tt.colorOutput,
			}

			got := formatter.Format(msg)

			if tt.want != "" {
				if got != tt.want {
					t.Errorf("TextFormatter.Format() = %q, want %q", got, tt.want)
				}
			}

			for _, wantStr := range tt.wantContains {
				if !strings.Contains(got, wantStr) {
					t.Errorf("TextFormatter.Format() = %q, want to contain %q", got, wantStr)
				}
			}
		})
	}
}
