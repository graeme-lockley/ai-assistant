package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ParseRequestBody reads the request body and extracts the user message based on contentType.
// Supported: application/json (body {"message":"..."}), text/plain (body is the message).
// Returns the message and nil, or empty string and error for unsupported type or read error.
func ParseRequestBody(r io.Reader, contentType string) (message string, err error) {
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch ct {
	case ContentTypeJSON:
		var req Request
		if err := json.NewDecoder(r).Decode(&req); err != nil {
			return "", fmt.Errorf("parse json body: %w", err)
		}
		return req.Message, nil
	case ContentTypeText:
		body, err := io.ReadAll(r)
		if err != nil {
			return "", fmt.Errorf("read body: %w", err)
		}
		return strings.TrimSpace(string(body)), nil
	default:
		return "", fmt.Errorf("unsupported content type: %q", contentType)
	}
}
