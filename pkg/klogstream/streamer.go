package klogstream

import (
	"context"
	"fmt"

	"github.com/archsyscall/klogstream/internal/filter"
	"github.com/archsyscall/klogstream/internal/kube"
	"github.com/archsyscall/klogstream/internal/stream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Streamer is the main interface for streaming logs
type Streamer interface {
	// Start begins streaming logs for matching pods
	Start(ctx context.Context) error
	// Stop stops all log streaming activity
	Stop()
}

// streamerImpl is the implementation of the Streamer interface
type streamerImpl struct {
	internal *stream.Streamer
}

// NewStreamer creates a new Streamer with the given options
var NewStreamer = func(options ...StreamOption) (Streamer, error) {
	// Create default config
	config := NewStreamConfig()

	// Apply options
	for _, option := range options {
		option(config)
	}

	// Convert to internal types
	internalFilter, err := convertFilter(config.Filter)
	if err != nil {
		return nil, err
	}

	// Create internal client provider
	clientProvider := kube.NewClientProviderWithOptions(config.KubeOptions...)

	// Create internal streamer config
	internalConfig := &stream.StreamerConfig{
		KubeClientProvider: clientProvider,
		Filter:             internalFilter,
		RetryPolicy: stream.RetryPolicy{
			MaxRetries:      config.RetryPolicy.MaxRetries,
			InitialInterval: config.RetryPolicy.InitialInterval,
			MaxInterval:     config.RetryPolicy.MaxInterval,
			Multiplier:      config.RetryPolicy.Multiplier,
		},
	}

	// Set handler with adapter
	if config.Handler != nil {
		internalConfig.Handler = stream.NewHandlerAdapter(adaptHandler(config.Handler))
	}

	// Set formatter with adapter if provided
	if config.Formatter != nil {
		internalConfig.Formatter = stream.NewFormatterAdapter(adaptFormatter(config.Formatter))
	}

	// Set matcher with adapter if provided
	if config.Matcher != nil {
		internalConfig.Matcher = stream.NewMatcherAdapter(adaptMatcher(config.Matcher))
	}

	// Create internal streamer
	internalStreamer, err := stream.NewStreamer(internalConfig)
	if err != nil {
		return nil, err
	}

	return &streamerImpl{
		internal: internalStreamer,
	}, nil
}

// Start begins streaming logs for matching pods
func (s *streamerImpl) Start(ctx context.Context) error {
	return s.internal.Start(ctx)
}

// Stop stops all log streaming activity
func (s *streamerImpl) Stop() {
	s.internal.Stop()
}

// convertFilter converts a public LogFilter to an internal filter
func convertFilter(logFilter *LogFilter) (*filter.LogFilter, error) {
	if logFilter == nil {
		return nil, fmt.Errorf("log filter is required")
	}

	f := &filter.LogFilter{
		PodNameRegex:   logFilter.PodNameRegex,
		ContainerRegex: logFilter.ContainerRegex,
		LabelSelector:  logFilter.LabelSelector,
		IncludeRegex:   logFilter.IncludeRegex,
		Since:          logFilter.Since,
		ContainerState: logFilter.ContainerState,
		Namespaces:     logFilter.Namespaces,
	}

	// Set default container state if not specified
	if f.ContainerState == "" {
		f.ContainerState = "all"
	}

	// Validate the filter
	if err := f.Validate(); err != nil {
		return nil, err
	}

	return f, nil
}

// handlerWrapper adapts the public LogHandler to the stream.ExternalLogHandler interface
type handlerWrapper struct {
	handler LogHandler
}

func (w *handlerWrapper) OnLog(msg interface{}) {
	if logMsg, ok := msg.(stream.LogMessage); ok {
		w.handler.OnLog(LogMessage{
			Namespace:     logMsg.Namespace,
			PodName:       logMsg.PodName,
			ContainerName: logMsg.ContainerName,
			Timestamp:     logMsg.Timestamp,
			Message:       logMsg.Message,
			Raw:           logMsg.Raw,
		})
	}
}

func (w *handlerWrapper) OnError(err error) {
	w.handler.OnError(err)
}

func (w *handlerWrapper) OnEnd() {
	w.handler.OnEnd()
}

// adaptHandler adapts the public LogHandler to the stream.ExternalLogHandler interface
func adaptHandler(handler LogHandler) stream.ExternalLogHandler {
	return &handlerWrapper{handler: handler}
}

// formatterWrapper adapts the public LogFormatter to the stream.ExternalLogFormatter interface
type formatterWrapper struct {
	formatter LogFormatter
}

func (w *formatterWrapper) Format(msg interface{}) string {
	if logMsg, ok := msg.(stream.LogMessage); ok {
		return w.formatter.Format(LogMessage{
			Namespace:     logMsg.Namespace,
			PodName:       logMsg.PodName,
			ContainerName: logMsg.ContainerName,
			Timestamp:     logMsg.Timestamp,
			Message:       logMsg.Message,
			Raw:           logMsg.Raw,
		})
	}
	return ""
}

// adaptFormatter adapts the public LogFormatter to the stream.ExternalLogFormatter interface
func adaptFormatter(formatter LogFormatter) stream.ExternalLogFormatter {
	return &formatterWrapper{formatter: formatter}
}

// matcherWrapper adapts the public MultilineMatcher to the stream.ExternalMatcher interface
type matcherWrapper struct {
	matcher MultilineMatcher
}

func (w *matcherWrapper) ShouldMerge(previous, next string) bool {
	return w.matcher.ShouldMerge(previous, next)
}

// adaptMatcher adapts the public MultilineMatcher to the stream.ExternalMatcher interface
func adaptMatcher(matcher MultilineMatcher) stream.ExternalMatcher {
	return &matcherWrapper{matcher: matcher}
}

// Run is a convenience function that creates a streamer with the given options,
// starts it, and waits for context completion
func Run(ctx context.Context, options ...StreamOption) error {
	streamer, err := NewStreamer(options...)
	if err != nil {
		return err
	}

	// Start streaming
	if err := streamer.Start(ctx); err != nil {
		return err
	}

	// Wait for context completion
	<-ctx.Done()

	// Stop streaming
	streamer.Stop()

	return nil
}

// StreamBuilder provides a fluent API for building and running a streamer
type StreamBuilder struct {
	options []StreamOption
}

// NewBuilder creates a new StreamBuilder
func NewBuilder() *StreamBuilder {
	return &StreamBuilder{}
}

// WithRestConfig adds a rest.Config option to the builder
func (b *StreamBuilder) WithRestConfig(config *rest.Config) *StreamBuilder {
	b.options = append(b.options, WithRestConfig(config))
	return b
}

// WithKubeconfigPath adds a kubeconfig path option to the builder
func (b *StreamBuilder) WithKubeconfigPath(path string) *StreamBuilder {
	b.options = append(b.options, WithKubeconfigPath(path))
	return b
}

// WithKubeContext adds a kubernetes context option to the builder
func (b *StreamBuilder) WithKubeContext(name string) *StreamBuilder {
	b.options = append(b.options, WithKubeContext(name))
	return b
}

// WithClientset adds a direct kubernetes clientset option to the builder
// This is especially useful for testing with fake.Clientset
func (b *StreamBuilder) WithClientset(clientset *kubernetes.Clientset) *StreamBuilder {
	b.options = append(b.options, WithClientset(clientset))
	return b
}

// WithNamespace adds a namespace to the log filter
func (b *StreamBuilder) WithNamespace(namespace string) *StreamBuilder {
	b.options = append(b.options, WithNamespace(namespace))
	return b
}

// WithPodRegex adds a pod name regex to the log filter
func (b *StreamBuilder) WithPodRegex(pattern string) *StreamBuilder {
	b.options = append(b.options, WithPodRegex(pattern))
	return b
}

// WithContainerRegex adds a container name regex to the log filter
func (b *StreamBuilder) WithContainerRegex(pattern string) *StreamBuilder {
	b.options = append(b.options, WithContainerRegex(pattern))
	return b
}

// WithLabel adds a label selector to the log filter
func (b *StreamBuilder) WithLabel(key, value string) *StreamBuilder {
	b.options = append(b.options, WithLabel(key, value))
	return b
}

// WithPodLabelSelector adds a label selector string to the log filter
// The format is the same as kubectl's label selector (e.g., "app=myapp,env=prod")
func (b *StreamBuilder) WithPodLabelSelector(selector string) *StreamBuilder {
	b.options = append(b.options, WithLabelSelector(selector))
	return b
}

// WithIncludeRegex adds an include regex to the log filter
func (b *StreamBuilder) WithIncludeRegex(pattern string) *StreamBuilder {
	b.options = append(b.options, WithIncludeRegex(pattern))
	return b
}

// WithFormatter sets the log formatter
func (b *StreamBuilder) WithFormatter(formatter LogFormatter) *StreamBuilder {
	b.options = append(b.options, WithFormatter(formatter))
	return b
}

// WithHandler sets the log handler
func (b *StreamBuilder) WithHandler(handler LogHandler) *StreamBuilder {
	b.options = append(b.options, WithHandler(handler))
	return b
}

// WithMatcher sets the multiline matcher
func (b *StreamBuilder) WithMatcher(matcher MultilineMatcher) *StreamBuilder {
	b.options = append(b.options, WithMatcher(matcher))
	return b
}

// Build creates a Streamer from the accumulated options
func (b *StreamBuilder) Build() (Streamer, error) {
	return NewStreamer(b.options...)
}

// Run creates a Streamer from the accumulated options, starts it, and waits for context completion
func (b *StreamBuilder) Run(ctx context.Context) error {
	return Run(ctx, b.options...)
}
