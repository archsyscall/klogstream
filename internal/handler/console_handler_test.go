package handler

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestConsoleHandler_OnLog(t *testing.T) {
	// Create a buffer to capture output
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	// Create handler with buffer
	handler := NewConsoleHandlerWithWriters(outBuf, errBuf)

	// Create a sample message
	msg := LogMessage{
		Namespace:     "default",
		PodName:       "test-pod",
		ContainerName: "test-container",
		Timestamp:     time.Now(),
		Message:       "Test message",
	}

	// Call OnLog
	handler.OnLog(msg)

	// Check output
	if outBuf.String() != "Test message\n" {
		t.Errorf("Expected 'Test message\\n', got %q", outBuf.String())
	}

	// Check no error output
	if errBuf.String() != "" {
		t.Errorf("Expected empty error buffer, got %q", errBuf.String())
	}
}

func TestConsoleHandler_OnError(t *testing.T) {
	// Create a buffer to capture output
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	// Create handler with buffer
	handler := NewConsoleHandlerWithWriters(outBuf, errBuf)

	// Create a sample error
	err := errors.New("test error")

	// Call OnError
	handler.OnError(err)

	// Check no standard output
	if outBuf.String() != "" {
		t.Errorf("Expected empty output buffer, got %q", outBuf.String())
	}

	// Check error output
	expected := "Error: test error\n"
	if errBuf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, errBuf.String())
	}
}

func TestConsoleHandler_OnEnd(t *testing.T) {
	// Create a buffer to capture output
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	// Create handler with buffer
	handler := NewConsoleHandlerWithWriters(outBuf, errBuf)

	// Call OnEnd
	handler.OnEnd()

	// Check no output
	if outBuf.String() != "" {
		t.Errorf("Expected empty output buffer, got %q", outBuf.String())
	}

	// Check no error output
	if errBuf.String() != "" {
		t.Errorf("Expected empty error buffer, got %q", errBuf.String())
	}
}
