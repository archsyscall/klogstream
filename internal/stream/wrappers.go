package stream

// HandlerWrapper is a wrapper for external LogHandler implementations
type HandlerWrapper struct {
	handler interface {
		OnLog(interface{})
		OnError(error)
		OnEnd()
	}
}

// NewHandlerWrapper creates a new HandlerWrapper
func NewHandlerWrapper(handler interface {
	OnLog(interface{})
	OnError(error)
	OnEnd()
}) *HandlerWrapper {
	return &HandlerWrapper{
		handler: handler,
	}
}

// OnLog handles a log message
func (w *HandlerWrapper) OnLog(msg LogMessage) {
	w.handler.OnLog(msg)
}

// OnError handles an error
func (w *HandlerWrapper) OnError(err error) {
	w.handler.OnError(err)
}

// OnEnd signals the end of streaming
func (w *HandlerWrapper) OnEnd() {
	w.handler.OnEnd()
}

// FormatterWrapper is a wrapper for external LogFormatter implementations
type FormatterWrapper struct {
	formatter interface {
		Format(interface{}) string
	}
}

// NewFormatterWrapper creates a new FormatterWrapper
func NewFormatterWrapper(formatter interface {
	Format(interface{}) string
}) *FormatterWrapper {
	return &FormatterWrapper{
		formatter: formatter,
	}
}

// Format formats a log message
func (w *FormatterWrapper) Format(msg LogMessage) string {
	return w.formatter.Format(msg)
}

// MatcherWrapper is a wrapper for external MultilineMatcher implementations
type MatcherWrapper struct {
	matcher interface {
		ShouldMerge(previous, next string) bool
	}
}

// NewMatcherWrapper creates a new MatcherWrapper
func NewMatcherWrapper(matcher interface {
	ShouldMerge(previous, next string) bool
}) *MatcherWrapper {
	return &MatcherWrapper{
		matcher: matcher,
	}
}

// ShouldMerge determines if the next line should be merged with the previous line
func (w *MatcherWrapper) ShouldMerge(previous, next string) bool {
	return w.matcher.ShouldMerge(previous, next)
}
