package klogstream

import (
	"testing"
	"time"
)

func TestLogFilterBuilder_Public(t *testing.T) {
	// Test the public builder API
	builder := NewLogFilterBuilder()

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

func TestLogFilterBuilder_EmptyBuild(t *testing.T) {
	// Test that building with no options fails
	builder := NewLogFilterBuilder()
	_, err := builder.Build()

	// Should get an error
	if err == nil {
		t.Errorf("LogFilterBuilder.Build() should have failed with empty filter")
	}
}
