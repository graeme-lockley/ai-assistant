package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ParseRequestBody reads the request body and extracts the user message and optional model based on contentType.
// Supported: application/json (body {"message":"...", "model":"..."}), text/plain (body is the message).
// Returns the message, model (may be empty), and error.
func ParseRequestBody(r io.Reader, contentType string) (message string, model string, err error) {
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch ct {
	case ContentTypeJSON:
		var req Request
		if err := json.NewDecoder(r).Decode(&req); err != nil {
			return "", "", fmt.Errorf("parse json body: %w", err)
		}
		return req.Message, req.Model, nil
	case ContentTypeText:
		body, err := io.ReadAll(r)
		if err != nil {
			return "", "", fmt.Errorf("read body: %w", err)
		}
		return strings.TrimSpace(string(body)), "", nil
	default:
		return "", "", fmt.Errorf("unsupported content type: %q", contentType)
	}
}
