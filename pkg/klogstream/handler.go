package klogstream

import (
	"io"

	"github.com/archsyscall/klogstream/internal/handler"
)

// ConsoleHandler outputs logs to the console
type ConsoleHandler struct {
	internal *handler.ConsoleHandler
}

// NewConsoleHandler creates a new ConsoleHandler with stdout and stderr as default outputs
func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{
		internal: handler.NewConsoleHandler(),
	}
}

// NewConsoleHandlerWithWriters creates a new ConsoleHandler with custom writers
func NewConsoleHandlerWithWriters(out, errOut io.Writer) *ConsoleHandler {
	return &ConsoleHandler{
		internal: handler.NewConsoleHandlerWithWriters(out, errOut),
	}
}

// OnLog writes formatted log messages to the configured output writer
func (h *ConsoleHandler) OnLog(msg LogMessage) {
	// Convert our LogMessage to the internal type
	internalMsg := handler.LogMessage{
		Namespace:     msg.Namespace,
		PodName:       msg.PodName,
		ContainerName: msg.ContainerName,
		Timestamp:     msg.Timestamp,
		Message:       msg.Message,
		Raw:           msg.Raw,
	}

	h.internal.OnLog(internalMsg)
}

// OnError writes error messages to the error output writer
func (h *ConsoleHandler) OnError(err error) {
	h.internal.OnError(err)
}

// OnEnd is called when the stream ends
func (h *ConsoleHandler) OnEnd() {
	h.internal.OnEnd()
}
