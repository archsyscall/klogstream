package formatter

import (
	"bytes"
	"text/template"
)

// TemplateFormatter formats log messages using Go templates
type TemplateFormatter struct {
	// Template is the parsed template for formatting
	Template *template.Template
}

// DefaultTemplate is the default template format
const DefaultTemplate = "{{.Timestamp}} [{{.Namespace}}] {{.PodName}}/{{.ContainerName}}: {{.Message}}"

// NewTemplateFormatter creates a new TemplateFormatter with the default template
func NewTemplateFormatter() (*TemplateFormatter, error) {
	return NewTemplateFormatterWithTemplate(DefaultTemplate)
}

// NewTemplateFormatterWithTemplate creates a new TemplateFormatter with a custom template
func NewTemplateFormatterWithTemplate(templateStr string) (*TemplateFormatter, error) {
	tmpl, err := template.New("log").Parse(templateStr)
	if err != nil {
		return nil, err
	}

	return &TemplateFormatter{
		Template: tmpl,
	}, nil
}

// Format converts a LogMessage to a formatted string using the template
func (f *TemplateFormatter) Format(msg LogMessage) string {
	var buf bytes.Buffer
	err := f.Template.Execute(&buf, msg)
	if err != nil {
		// Fallback in case of template execution error
		return msg.Message
	}

	return buf.String()
}
