package stream

// ExternalLogHandler is an interface that represents external log handlers
type ExternalLogHandler interface {
	OnLog(interface{})
	OnError(error)
	OnEnd()
}

// ExternalLogFormatter is an interface that represents external log formatters
type ExternalLogFormatter interface {
	Format(interface{}) string
}

// ExternalMatcher is an interface that represents external multiline matchers
type ExternalMatcher interface {
	ShouldMerge(previous, next string) bool
}

// HandlerAdapter adapts internal LogMessage to external handlers
type HandlerAdapter struct {
	ExternalHandler ExternalLogHandler
}

// NewHandlerAdapter creates a new HandlerAdapter
func NewHandlerAdapter(handler ExternalLogHandler) *HandlerAdapter {
	return &HandlerAdapter{
		ExternalHandler: handler,
	}
}

// OnLog forwards the log message to the external handler
func (a *HandlerAdapter) OnLog(msg LogMessage) {
	a.ExternalHandler.OnLog(msg)
}

// OnError forwards the error to the external handler
func (a *HandlerAdapter) OnError(err error) {
	a.ExternalHandler.OnError(err)
}

// OnEnd forwards the end signal to the external handler
func (a *HandlerAdapter) OnEnd() {
	a.ExternalHandler.OnEnd()
}

// FormatterAdapter adapts internal LogMessage to external formatters
type FormatterAdapter struct {
	ExternalFormatter ExternalLogFormatter
}

// NewFormatterAdapter creates a new FormatterAdapter
func NewFormatterAdapter(formatter ExternalLogFormatter) *FormatterAdapter {
	return &FormatterAdapter{
		ExternalFormatter: formatter,
	}
}

// Format forwards the log message to the external formatter
func (a *FormatterAdapter) Format(msg LogMessage) string {
	return a.ExternalFormatter.Format(msg)
}

// MatcherAdapter adapts external multiline matchers to internal interface
type MatcherAdapter struct {
	ExternalMatcher ExternalMatcher
}

// NewMatcherAdapter creates a new MatcherAdapter
func NewMatcherAdapter(matcher ExternalMatcher) *MatcherAdapter {
	return &MatcherAdapter{
		ExternalMatcher: matcher,
	}
}

// ShouldMerge forwards the call to the external matcher
func (a *MatcherAdapter) ShouldMerge(previous, next string) bool {
	return a.ExternalMatcher.ShouldMerge(previous, next)
}
