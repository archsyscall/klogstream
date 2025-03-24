package klogstream

import (
	"github.com/archsyscall/klogstream/internal/formatter"
)

// TextFormatter formats log messages as text
type TextFormatter struct {
	// ShowTimestamp controls whether to display the timestamp
	ShowTimestamp bool
	// ShowNamespace controls whether to display the namespace
	ShowNamespace bool
	// ShowPodName controls whether to display the pod name
	ShowPodName bool
	// ShowContainerName controls whether to display the container name
	ShowContainerName bool
	// TimestampFormat defines the format for timestamps
	TimestampFormat string
	// ColorOutput enables colorized output
	ColorOutput bool

	internal *formatter.TextFormatter
}

// NewTextFormatter creates a new TextFormatter with default settings
func NewTextFormatter() *TextFormatter {
	internal := formatter.NewTextFormatter()
	return &TextFormatter{
		ShowTimestamp:     internal.ShowTimestamp,
		ShowNamespace:     internal.ShowNamespace,
		ShowPodName:       internal.ShowPodName,
		ShowContainerName: internal.ShowContainerName,
		TimestampFormat:   internal.TimestampFormat,
		ColorOutput:       internal.ColorOutput,
		internal:          internal,
	}
}

// Format converts a LogMessage to a formatted string
func (f *TextFormatter) Format(msg LogMessage) string {
	// Update internal formatter with current settings
	f.internal.ShowTimestamp = f.ShowTimestamp
	f.internal.ShowNamespace = f.ShowNamespace
	f.internal.ShowPodName = f.ShowPodName
	f.internal.ShowContainerName = f.ShowContainerName
	f.internal.TimestampFormat = f.TimestampFormat
	f.internal.ColorOutput = f.ColorOutput

	// Convert our LogMessage to the internal type
	internalMsg := formatter.LogMessage{
		Namespace:     msg.Namespace,
		PodName:       msg.PodName,
		ContainerName: msg.ContainerName,
		Timestamp:     msg.Timestamp,
		Message:       msg.Message,
		Raw:           msg.Raw,
	}

	return f.internal.Format(internalMsg)
}

// JSONFormatter formats log messages as JSON
type JSONFormatter struct {
	// IncludeTimestamp controls whether to include the timestamp in the JSON
	IncludeTimestamp bool
	// IncludeNamespace controls whether to include the namespace in the JSON
	IncludeNamespace bool
	// IncludePodName controls whether to include the pod name in the JSON
	IncludePodName bool
	// IncludeContainerName controls whether to include the container name in the JSON
	IncludeContainerName bool

	internal *formatter.JSONFormatter
}

// NewJSONFormatter creates a new JSONFormatter with default settings
func NewJSONFormatter() *JSONFormatter {
	internal := formatter.NewJSONFormatter()
	return &JSONFormatter{
		IncludeTimestamp:     internal.IncludeTimestamp,
		IncludeNamespace:     internal.IncludeNamespace,
		IncludePodName:       internal.IncludePodName,
		IncludeContainerName: internal.IncludeContainerName,
		internal:             internal,
	}
}

// Format converts a LogMessage to a JSON string
func (f *JSONFormatter) Format(msg LogMessage) string {
	// Update internal formatter with current settings
	f.internal.IncludeTimestamp = f.IncludeTimestamp
	f.internal.IncludeNamespace = f.IncludeNamespace
	f.internal.IncludePodName = f.IncludePodName
	f.internal.IncludeContainerName = f.IncludeContainerName

	// Convert our LogMessage to the internal type
	internalMsg := formatter.LogMessage{
		Namespace:     msg.Namespace,
		PodName:       msg.PodName,
		ContainerName: msg.ContainerName,
		Timestamp:     msg.Timestamp,
		Message:       msg.Message,
		Raw:           msg.Raw,
	}

	return f.internal.Format(internalMsg)
}

// TemplateFormatter formats log messages using Go templates
type TemplateFormatter struct {
	// TemplateString is the template string to use
	TemplateString string

	internal *formatter.TemplateFormatter
}

// NewTemplateFormatter creates a new TemplateFormatter with the default template
func NewTemplateFormatter() (*TemplateFormatter, error) {
	internal, err := formatter.NewTemplateFormatter()
	if err != nil {
		return nil, err
	}

	return &TemplateFormatter{
		TemplateString: formatter.DefaultTemplate,
		internal:       internal,
	}, nil
}

// NewTemplateFormatterWithTemplate creates a new TemplateFormatter with a custom template
func NewTemplateFormatterWithTemplate(templateStr string) (*TemplateFormatter, error) {
	internal, err := formatter.NewTemplateFormatterWithTemplate(templateStr)
	if err != nil {
		return nil, err
	}

	return &TemplateFormatter{
		TemplateString: templateStr,
		internal:       internal,
	}, nil
}

// Format converts a LogMessage to a formatted string using the template
func (f *TemplateFormatter) Format(msg LogMessage) string {
	// Convert our LogMessage to the internal type
	internalMsg := formatter.LogMessage{
		Namespace:     msg.Namespace,
		PodName:       msg.PodName,
		ContainerName: msg.ContainerName,
		Timestamp:     msg.Timestamp,
		Message:       msg.Message,
		Raw:           msg.Raw,
	}

	return f.internal.Format(internalMsg)
}
