package stream

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/archsyscall/klogstream/internal/errors"
	"github.com/archsyscall/klogstream/internal/filter"
	"github.com/archsyscall/klogstream/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LogHandler is an interface for handling log messages and errors
type LogHandler interface {
	OnLog(LogMessage)
	OnError(error)
	OnEnd()
}

// LogFormatter is an interface for formatting log messages
type LogFormatter interface {
	Format(LogMessage) string
}

// MultilineMatcher is an interface for matching multiline log entries
type MultilineMatcher interface {
	ShouldMerge(previous, next string) bool
}

// RetryPolicy configures the retry behavior for transient errors
type RetryPolicy struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// LogMessage represents a single log entry from a kubernetes pod/container
type LogMessage struct {
	Namespace     string
	PodName       string
	ContainerName string
	Timestamp     time.Time
	Message       string
	Raw           []byte
}

// LogStreamError represents an error that occurred during log streaming
type LogStreamError struct {
	Err       error
	Permanent bool
	Reason    string
}

// Error implements the error interface
func (e *LogStreamError) Error() string {
	if e.Reason != "" {
		return e.Reason + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *LogStreamError) Unwrap() error {
	return e.Err
}

// NewLogStreamError creates a new LogStreamError
func NewLogStreamError(err error, permanent bool, reason string) *LogStreamError {
	return &LogStreamError{
		Err:       err,
		Permanent: permanent,
		Reason:    reason,
	}
}

// Streamer handles streaming logs from multiple pods
type Streamer struct {
	clientset     *kubernetes.Clientset
	filter        *filter.LogFilter
	handler       LogHandler
	formatter     LogFormatter
	matcher       MultilineMatcher
	retryPolicy   RetryPolicy
	maxMultilines int
	active        sync.Map
	stopped       bool
	stopOnce      sync.Once
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// StreamerConfig contains configuration for the streamer
type StreamerConfig struct {
	KubeClientProvider *kube.ClientProvider
	Filter             *filter.LogFilter
	Handler            LogHandler
	Formatter          LogFormatter
	Matcher            MultilineMatcher
	RetryPolicy        RetryPolicy
	MaxMultilines      int
}

// DefaultMaxMultilines is the default maximum number of lines in a multiline log
const DefaultMaxMultilines = 500

// NewStreamer creates a new Streamer with the provided configuration
func NewStreamer(config *StreamerConfig) (*Streamer, error) {
	if config.KubeClientProvider == nil {
		return nil, fmt.Errorf("kubernetes client provider is required")
	}

	clientset, err := config.KubeClientProvider.GetClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	if config.Filter == nil {
		return nil, fmt.Errorf("log filter is required")
	}

	if config.Handler == nil {
		return nil, fmt.Errorf("log handler is required")
	}

	// Set default formatter if not provided
	formatter := config.Formatter
	if formatter == nil {
		// Using a simple passthrough formatter
		formatter = &passthroughFormatter{}
	}

	// Set default max multilines if not provided
	maxMultilines := config.MaxMultilines
	if maxMultilines <= 0 {
		maxMultilines = DefaultMaxMultilines
	}

	return &Streamer{
		clientset:     clientset,
		filter:        config.Filter,
		handler:       config.Handler,
		formatter:     formatter,
		matcher:       config.Matcher,
		retryPolicy:   config.RetryPolicy,
		maxMultilines: maxMultilines,
		stopCh:        make(chan struct{}),
	}, nil
}

// passthrough formatter just returns the message as is
type passthroughFormatter struct{}

func (f *passthroughFormatter) Format(msg LogMessage) string {
	return msg.Message
}

// Start begins streaming logs for matching pods
func (s *Streamer) Start(ctx context.Context) error {
	// Check if already stopped
	if s.stopped {
		return NewLogStreamError(fmt.Errorf("streamer is stopped"), true, "streamer stopped")
	}

	// Create a context that can be canceled when Stop is called
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-s.stopCh:
			cancel()
		case <-ctx.Done():
			// Context was canceled elsewhere
		}
	}()

	// Start the pod watcher to continuously watch for matching pods
	return s.startPodWatcher(ctx)
}

// Stop stops all log streaming activity
func (s *Streamer) Stop() {
	s.stopOnce.Do(func() {
		s.stopped = true
		close(s.stopCh)
		s.wg.Wait()
		s.handler.OnEnd()
	})
}

// startPodWatcher starts a goroutine to watch for pods matching the filter
func (s *Streamer) startPodWatcher(ctx context.Context) error {
	// Start a watcher for each namespace
	for _, namespace := range s.filter.Namespaces {
		// Create watch for pods in this namespace
		labelSelector := ""
		if s.filter.LabelSelector != nil {
			labelSelector = s.filter.LabelSelector.String()
		}

		// Start by listing existing pods
		pods, err := s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return NewLogStreamError(err, true, "failed to list pods")
		}

		// Start streaming logs for existing pods
		for _, pod := range pods.Items {
			if s.shouldStreamPod(&pod) {
				s.startPodLogStreamer(ctx, &pod)
			}
		}

		// Now watch for new pods
		s.wg.Add(1)
		go func(ns string) {
			defer s.wg.Done()

			// Use a retry loop for the watcher
			retry := 0
			backoff := s.retryPolicy.InitialInterval

			for {
				// Check if we should stop
				select {
				case <-ctx.Done():
					return
				case <-s.stopCh:
					return
				default:
					// Continue
				}

				// Create a watch for pods
				watcher, err := s.clientset.CoreV1().Pods(ns).Watch(ctx, metav1.ListOptions{
					LabelSelector: labelSelector,
					// Ignore too old events by setting the resource version
					ResourceVersion: "0",
					// Timeout after a while so we can check for cancellation
					TimeoutSeconds: new(int64),
				})

				if err != nil {
					// Check if this is a permanent error
					if isPermError(err) {
						s.handler.OnError(NewLogStreamError(err, true, "failed to watch pods"))
						return
					}

					// Handle transient error
					s.handler.OnError(NewLogStreamError(err, false, "failed to watch pods"))

					// Retry with backoff
					retry++
					if retry > s.retryPolicy.MaxRetries {
						s.handler.OnError(NewLogStreamError(fmt.Errorf("exceeded maximum retries"), true, "pod watch retries exceeded"))
						return
					}

					// Sleep with backoff
					select {
					case <-time.After(backoff):
						// Increase backoff for next retry
						backoff = time.Duration(float64(backoff) * s.retryPolicy.Multiplier)
						if backoff > s.retryPolicy.MaxInterval {
							backoff = s.retryPolicy.MaxInterval
						}
					case <-ctx.Done():
						return
					case <-s.stopCh:
						return
					}

					continue
				}

				// Reset retry counter on successful watch
				retry = 0
				backoff = s.retryPolicy.InitialInterval

				// Process events
				for event := range watcher.ResultChan() {
					// Check if we should stop
					select {
					case <-ctx.Done():
						watcher.Stop()
						return
					case <-s.stopCh:
						watcher.Stop()
						return
					default:
						// Continue
					}

					// Process the pod event
					switch event.Type {
					case "ADDED", "MODIFIED":
						if pod, ok := event.Object.(*corev1.Pod); ok {
							if s.shouldStreamPod(pod) {
								// Check if we're already streaming this pod
								if _, exists := s.active.Load(pod.Name); !exists {
									s.startPodLogStreamer(ctx, pod)
								}
							}

							// Check if pod has completed (Succeeded or Failed phase)
							if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
								// Pod has completed, stop tracking it
								s.active.Delete(pod.Name)
							}
						}
					case "DELETED":
						if pod, ok := event.Object.(*corev1.Pod); ok {
							// Pod is gone, stop any active streamers
							s.active.Delete(pod.Name)
						}
					}
				}

				// If we get here, the watch channel was closed, retry
			}
		}(namespace)
	}

	return nil
}

// shouldStreamPod checks if a pod matches the filter criteria
func (s *Streamer) shouldStreamPod(pod *corev1.Pod) bool {
	// Check pod name regex if specified
	if s.filter.PodNameRegex != nil && !s.filter.PodNameRegex.MatchString(pod.Name) {
		return false
	}

	// Always match at the pod level even if we filter at the container level
	return true
}

// startPodLogStreamer starts a goroutine to stream logs for each matching container in the pod
func (s *Streamer) startPodLogStreamer(ctx context.Context, pod *corev1.Pod) {
	// Mark this pod as active
	s.active.Store(pod.Name, true)

	// Start a streamer for each container that matches
	for _, container := range pod.Spec.Containers {
		// Check container name regex if specified
		if s.filter.ContainerRegex != nil && !s.filter.ContainerRegex.MatchString(container.Name) {
			continue
		}

		// Check container state if specified
		if s.filter.ContainerState != "all" {
			// TODO: Implement container state filtering
			// For now we always stream
		}

		// Start the container log streamer
		s.wg.Add(1)
		go func(podName, containerName, namespace string) {
			defer s.wg.Done()

			// Use a retry loop for the log streaming
			retry := 0
			backoff := s.retryPolicy.InitialInterval

			for {
				// Check if we should stop
				select {
				case <-ctx.Done():
					return
				case <-s.stopCh:
					return
				default:
					// Continue
				}

				// Create the log options
				opts := &corev1.PodLogOptions{
					Container: containerName,
					Follow:    true,
				}

				// Set the since time if specified
				if s.filter.Since != nil {
					sinceTime := metav1.NewTime(*s.filter.Since)
					opts.SinceTime = &sinceTime
				}

				// Start streaming logs
				req := s.clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
				stream, err := req.Stream(ctx)
				if err != nil {
					// Check if this is a permanent error
					if isPermError(err) {
						s.handler.OnError(NewLogStreamError(err, true,
							fmt.Sprintf("failed to stream logs for pod %s container %s", podName, containerName)))
						return
					}

					// Handle transient error
					s.handler.OnError(NewLogStreamError(err, false,
						fmt.Sprintf("failed to stream logs for pod %s container %s", podName, containerName)))

					// Retry with backoff
					retry++
					if retry > s.retryPolicy.MaxRetries {
						s.handler.OnError(NewLogStreamError(fmt.Errorf("exceeded maximum retries"), true,
							fmt.Sprintf("log stream retries exceeded for pod %s container %s", podName, containerName)))
						return
					}

					// Sleep with backoff
					select {
					case <-time.After(backoff):
						// Increase backoff for next retry
						backoff = time.Duration(float64(backoff) * s.retryPolicy.Multiplier)
						if backoff > s.retryPolicy.MaxInterval {
							backoff = s.retryPolicy.MaxInterval
						}
					case <-ctx.Done():
						return
					case <-s.stopCh:
						return
					}

					continue
				}

				// Reset retry counter on successful stream
				retry = 0
				backoff = s.retryPolicy.InitialInterval

				// Process the log stream
				err = s.processLogStream(ctx, stream, podName, containerName, namespace)

				// Close the stream
				stream.Close()

				// If context canceled or stopped, exit
				select {
				case <-ctx.Done():
					return
				case <-s.stopCh:
					return
				default:
					// Continue
				}

				// If there was an error, decide whether to retry
				if err != nil {
					// Check if this is a permanent error
					if lse, ok := err.(*LogStreamError); ok && lse.Permanent {
						s.handler.OnError(lse)
						return
					}

					// Handle transient error
					s.handler.OnError(err)

					// Sleep with backoff before retrying
					select {
					case <-time.After(backoff):
						// Increase backoff for next retry
						backoff = time.Duration(float64(backoff) * s.retryPolicy.Multiplier)
						if backoff > s.retryPolicy.MaxInterval {
							backoff = s.retryPolicy.MaxInterval
						}
					case <-ctx.Done():
						return
					case <-s.stopCh:
						return
					}
				}
			}
		}(pod.Name, container.Name, pod.Namespace)
	}
}

// processLogStream reads log lines from the stream and processes them
func (s *Streamer) processLogStream(ctx context.Context, stream io.ReadCloser, podName, containerName, namespace string) error {
	// If we have a multiline matcher, use buffering logic
	if s.matcher != nil {
		return s.processMultilineLogStream(ctx, stream, podName, containerName, namespace)
	}

	// Simple single-line processing
	scanner := NewScanner(stream)
	for scanner.Scan() {
		// Check if we should stop
		select {
		case <-ctx.Done():
			return nil
		case <-s.stopCh:
			return nil
		default:
			// Continue
		}

		line := scanner.Text()

		// Check include regex if specified
		if s.filter.IncludeRegex != nil && !s.filter.IncludeRegex.MatchString(line) {
			continue
		}

		// Create the log message
		timestamp := time.Now() // Ideally we'd parse from the log line if possible
		msg := LogMessage{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			Timestamp:     timestamp,
			Message:       line,
			Raw:           scanner.Bytes(),
		}

		// Format the message
		msg.Message = s.formatter.Format(msg)

		// Send to handler
		s.handler.OnLog(msg)
	}

	if err := scanner.Err(); err != nil {
		// Check if this is a pod deletion error (normal termination)
		if errors.IsPodDeletedError(err) {
			// Pod deleted, remove from active tracking
			s.active.Delete(podName)
			// Just return nil for normal pod termination
			return nil
		}
		// Check if this is a permanent error
		if isPermError(err) {
			return NewLogStreamError(err, true, "log stream read error")
		}
		return NewLogStreamError(err, false, "log stream read error")
	}

	// End of stream, not an error
	return nil
}

// processMultilineLogStream reads log lines from the stream and processes them with multiline support
func (s *Streamer) processMultilineLogStream(ctx context.Context, stream io.ReadCloser, podName, containerName, namespace string) error {
	scanner := NewScanner(stream)

	var buffer []string
	var rawBuffer [][]byte
	var lastLine string

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		// Join the buffer
		message := buffer[0]
		for i := 1; i < len(buffer); i++ {
			message += "\n" + buffer[i]
		}

		// Check include regex if specified
		if s.filter.IncludeRegex != nil && !s.filter.IncludeRegex.MatchString(message) {
			// Reset buffer
			buffer = nil
			rawBuffer = nil
			return
		}

		// Create the log message
		timestamp := time.Now() // Ideally we'd parse from the log line if possible

		// Combine raw bytes
		var rawBytes []byte
		for i, raw := range rawBuffer {
			if i > 0 {
				rawBytes = append(rawBytes, '\n')
			}
			rawBytes = append(rawBytes, raw...)
		}

		msg := LogMessage{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			Timestamp:     timestamp,
			Message:       message,
			Raw:           rawBytes,
		}

		// Format the message
		msg.Message = s.formatter.Format(msg)

		// Send to handler
		s.handler.OnLog(msg)

		// Reset buffer
		buffer = nil
		rawBuffer = nil
	}

	for scanner.Scan() {
		// Check if we should stop
		select {
		case <-ctx.Done():
			return nil
		case <-s.stopCh:
			return nil
		default:
			// Continue
		}

		line := scanner.Text()

		// Handle first line
		if len(buffer) == 0 {
			buffer = append(buffer, line)
			rawBuffer = append(rawBuffer, scanner.Bytes())
			lastLine = line
			continue
		}

		// Check if we should merge this line
		if s.matcher.ShouldMerge(lastLine, line) {
			// Add to buffer
			buffer = append(buffer, line)
			rawBuffer = append(rawBuffer, scanner.Bytes())
			lastLine = line

			// Check if we've exceeded max lines
			if len(buffer) >= s.maxMultilines {
				// Flush the buffer
				flush()
			}
		} else {
			// Flush the previous buffer
			flush()

			// Start a new buffer
			buffer = append(buffer, line)
			rawBuffer = append(rawBuffer, scanner.Bytes())
			lastLine = line
		}
	}

	// Flush any remaining buffer
	flush()

	if err := scanner.Err(); err != nil {
		// Check if this is a pod deletion error (normal termination)
		if errors.IsPodDeletedError(err) {
			// Pod deleted, remove from active tracking
			s.active.Delete(podName)
			// Just return nil for normal pod termination
			return nil
		}
		// Check if this is a permanent error
		if isPermError(err) {
			return NewLogStreamError(err, true, "log stream read error")
		}
		return NewLogStreamError(err, false, "log stream read error")
	}

	// End of stream, not an error
	return nil
}

// isPermError checks if an error should be considered permanent
func isPermError(err error) bool {
	// TODO: Implement better detection of permanent errors
	return false
}

// NewScanner creates a new scanner for reading log lines
func NewScanner(r io.Reader) *scanner {
	return &scanner{
		reader: r,
		buf:    make([]byte, 4096),
	}
}

// scanner is a simple line scanner similar to bufio.Scanner but with more control
type scanner struct {
	reader io.Reader
	buf    []byte
	token  []byte
	err    error
}

// Scan advances the scanner to the next token
func (s *scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	var token []byte
	for {
		n, err := s.reader.Read(s.buf)
		if n > 0 {
			// Find newline
			for i := 0; i < n; i++ {
				if s.buf[i] == '\n' {
					token = append(token, s.buf[:i]...)
					s.token = token

					// Handle remaining data
					// TODO: Properly handle remaining data
					return true
				}
			}

			// No newline found, append all and continue
			token = append(token, s.buf[:n]...)
		}

		if err != nil {
			s.err = err
			if err == io.EOF {
				// Return last token if any
				if len(token) > 0 {
					s.token = token
					return true
				}
			}
			return false
		}
	}
}

// Text returns the current token as a string
func (s *scanner) Text() string {
	return string(s.token)
}

// Bytes returns the current token as bytes
func (s *scanner) Bytes() []byte {
	return s.token
}

// Err returns the last error encountered
func (s *scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
