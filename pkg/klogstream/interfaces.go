package klogstream

// LogHandler handles log messages and errors
type LogHandler interface {
	// OnLog is called for each log message
	OnLog(LogMessage)
	// OnError is called when an error occurs
	OnError(error)
	// OnEnd is called when log streaming ends
	OnEnd()
}

// LogFormatter formats log messages as strings
type LogFormatter interface {
	// Format converts a log message to a formatted string
	Format(LogMessage) string
}

// MultilineMatcher determines if log lines should be merged
type MultilineMatcher interface {
	// ShouldMerge returns true if the next line should be merged with the previous
	ShouldMerge(previous, next string) bool
}
