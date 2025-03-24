package formatter

import (
	"testing"
	"time"
)

func TestTemplateFormatter_Format(t *testing.T) {
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
		name        string
		templateStr string
		want        string
		expectError bool
	}{
		{
			name:        "default template",
			templateStr: DefaultTemplate,
			want:        "2023-04-15 12:34:56 +0000 UTC [default] test-pod/test-container: Test message",
		},
		{
			name:        "custom template",
			templateStr: "{{.PodName}} - {{.Message}}",
			want:        "test-pod - Test message",
		},
		{
			name:        "timestamp format",
			templateStr: "{{.Timestamp.Format \"2006-01-02\"}} {{.Message}}",
			want:        "2023-04-15 Test message",
		},
		{
			name:        "invalid template",
			templateStr: "{{.Invalid}",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var formatter *TemplateFormatter
			var err error

			if tt.templateStr == DefaultTemplate {
				formatter, err = NewTemplateFormatter()
			} else {
				formatter, err = NewTemplateFormatterWithTemplate(tt.templateStr)
			}

			if tt.expectError {
				if err == nil {
					t.Fatalf("NewTemplateFormatterWithTemplate(%q) expected error", tt.templateStr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewTemplateFormatterWithTemplate(%q) error = %v", tt.templateStr, err)
			}

			got := formatter.Format(msg)

			if got != tt.want {
				t.Errorf("TemplateFormatter.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateFormatter_FormatError(t *testing.T) {
	// Create a valid formatter with a template that will fail at execution time
	formatter, err := NewTemplateFormatterWithTemplate("{{.Timestamp.Invalid}}")
	if err != nil {
		t.Fatalf("NewTemplateFormatterWithTemplate() error = %v", err)
	}

	// Create a sample message
	msg := LogMessage{
		Message: "Test message",
	}

	// Execute should fail but we should get the message as fallback
	got := formatter.Format(msg)

	if got != msg.Message {
		t.Errorf("TemplateFormatter.Format() on error = %q, want %q", got, msg.Message)
	}
}
