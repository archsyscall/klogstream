package filter

import (
	"testing"
	"time"
)

func TestLogFilterBuilder_Build(t *testing.T) {
	tests := []struct {
		name      string
		buildFunc func(*LogFilterBuilder) *LogFilterBuilder
		wantErr   bool
	}{
		{
			name: "empty builder",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				return b
			},
			wantErr: true,
		},
		{
			name: "only namespace",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				return b.Namespace("default")
			},
			wantErr: false,
		},
		{
			name: "invalid regex",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				// This should silently handle the invalid regex and not set it
				return b.PodRegex("[invalid").Namespace("default")
			},
			wantErr: false,
		},
		{
			name: "negative duration",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				// This should silently handle the negative duration and not set it
				return b.Since(-10 * time.Minute).Namespace("default")
			},
			wantErr: false,
		},
		{
			name: "full configuration",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				return b.
					PodRegex("nginx-").
					ContainerRegex("web").
					Label("app", "web").
					Include("ERROR").
					Since(30 * time.Minute).
					ContainerState("running").
					Namespace("default")
			},
			wantErr: false,
		},
		{
			name: "invalid container state",
			buildFunc: func(b *LogFilterBuilder) *LogFilterBuilder {
				return b.
					PodRegex("nginx-").
					Namespace("default").
					ContainerState("invalid")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewLogFilterBuilder()
			builder = tt.buildFunc(builder)
			_, err := builder.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogFilterBuilder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestLogFilterBuilder_Chaining(t *testing.T) {
	builder := NewLogFilterBuilder()

	// Test chaining methods
	filter, err := builder.
		PodRegex("nginx-").
		ContainerRegex("web").
		Label("app", "web").
		Include("ERROR").
		Since(30 * time.Minute).
		ContainerState("running").
		Namespace("default").
		Build()

	// Verify no error
	if err != nil {
		t.Fatalf("LogFilterBuilder.Build() unexpected error: %v", err)
	}

	// Verify filter values
	if filter.PodNameRegex == nil || filter.PodNameRegex.String() != "nginx-" {
		t.Errorf("PodNameRegex not set correctly, got %v", filter.PodNameRegex)
	}

	if filter.ContainerRegex == nil || filter.ContainerRegex.String() != "web" {
		t.Errorf("ContainerRegex not set correctly, got %v", filter.ContainerRegex)
	}

	if filter.LabelSelector == nil {
		t.Errorf("LabelSelector not set correctly, got nil")
	}

	if filter.IncludeRegex == nil || filter.IncludeRegex.String() != "ERROR" {
		t.Errorf("IncludeRegex not set correctly, got %v", filter.IncludeRegex)
	}

	if filter.Since == nil {
		t.Errorf("Since not set correctly, got nil")
	}

	if filter.ContainerState != "running" {
		t.Errorf("ContainerState not set correctly, got %s", filter.ContainerState)
	}

	if len(filter.Namespaces) != 1 || filter.Namespaces[0] != "default" {
		t.Errorf("Namespaces not set correctly, got %v", filter.Namespaces)
	}
}
