package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

// HTTP stream content types and header name.
const (
	HeaderSessionID     = "X-Session-Id"
	HeaderSessionClose  = "X-Session-Close"
	ContentTypeJSON     = "application/json"
	ContentTypeText     = "text/plain"
	ContentTypeSSE      = "text/event-stream"
	ContentTypeNDJSON   = "application/json" // streamed as NDJSON; Accept may match "application/json"
	AcceptHeaderSSE     = "text/event-stream"
	AcceptHeaderNDJSON  = "application/json"
)

// Stream event types for SSE and NDJSON.
const (
	EventSession = "session"
	EventToken   = "token"
	EventDone    = "done"
	EventError   = "error"
)

// SSEWriter writes Server-Sent Events to w. Not safe for concurrent use.
type SSEWriter struct {
	w io.Writer
}

// NewSSEWriter returns an SSE writer that writes to w.
func NewSSEWriter(w io.Writer) *SSEWriter {
	return &SSEWriter{w: w}
}

// WriteEvent writes one SSE event: event name and optional data (JSON-encoded if obj is non-nil).
func (s *SSEWriter) WriteEvent(event string, obj interface{}) error {
	if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
		return err
	}
	if obj != nil {
		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("sse encode: %w", err)
		}
		if _, err := fmt.Fprintf(s.w, "data: %s\n", data); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(s.w, "\n")
	return err
}

// NDJSONWriter writes newline-delimited JSON objects to w. Not safe for concurrent use.
type NDJSONWriter struct {
	w io.Writer
}

// NewNDJSONWriter returns an NDJSON writer that writes to w.
func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	return &NDJSONWriter{w: w}
}

// WriteLine writes one JSON object as a single line (no newline inside the object).
func (n *NDJSONWriter) WriteLine(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("ndjson encode: %w", err)
	}
	_, err = fmt.Fprintf(n.w, "%s\n", data)
	return err
}

// StreamEvent represents a single stream event for NDJSON.
type StreamEvent struct {
	Type      string `json:"type"` // "session", "token", "done", "error"
	SessionID string `json:"session_id,omitempty"`
	Delta     string `json:"delta,omitempty"`
	Error     string `json:"error,omitempty"`
}
