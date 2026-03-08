package ask

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/protocol"
)

type Result struct {
	Entries   []MessageEntry `json:"entries"`
	Model     string         `json:"model"`
	SessionID string         `json:"session_id"`
	Tokens    int            `json:"tokens"`
}

type MessageEntry struct {
	Type    string `json:"type"` // "thinking" or "output"
	Content string `json:"content"`
}

type ErrorResult struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func Run(ctx context.Context, cfg config.Ask, instruction string) (Result, error) {
	baseURL := cfg.ServerURL
	if baseURL == "" {
		baseURL = "http://" + config.DefaultServerAddr
	}

	client := &http.Client{}

	accept := protocol.AcceptHeaderSSE
	if cfg.DefaultResponseType != "" {
		accept = cfg.DefaultResponseType
	}
	contentType := protocol.ContentTypeJSON
	if cfg.DefaultRequestType != "" {
		contentType = cfg.DefaultRequestType
	}

	var body io.Reader
	if contentType == protocol.ContentTypeJSON {
		payload, _ := json.Marshal(&protocol.Request{
			Message: instruction,
			Model:   cfg.Model,
		})
		body = bytes.NewReader(payload)
	} else {
		body = strings.NewReader(instruction)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/", body)
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get(protocol.HeaderSessionID)

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return Result{}, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(errBody))
	}

	var result Result
	result.SessionID = sessionID
	result.Model = cfg.Model

	useSSE := strings.Contains(accept, "event-stream")
	useNDJSON := strings.Contains(accept, "application/json") && !useSSE

	if !useSSE && !useNDJSON {
		useSSE = true
	}

	if useSSE {
		result.Entries, result.Tokens, err = consumeSSE(resp.Body)
	} else {
		result.Entries, result.Tokens, err = consumeNDJSON(resp.Body)
	}
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}

	closeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/", nil)
	if err != nil {
		return Result{}, fmt.Errorf("build close request: %w", err)
	}
	closeReq.Header.Set(protocol.HeaderSessionID, sessionID)
	closeReq.Header.Set(protocol.HeaderSessionClose, "true")
	client.Do(closeReq)

	return result, nil
}

func consumeSSE(r io.Reader) (entries []MessageEntry, tokens int, err error) {
	scanner := bufio.NewScanner(r)
	var eventType string
	var data strings.Builder
	var currentThinking strings.Builder
	var currentOutput strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if eventType != "" {
				dataStr := strings.TrimSpace(data.String())
				switch eventType {
				case protocol.EventThinking:
					if dataStr != "" {
						var ev struct {
							Delta string `json:"delta"`
						}
						if json.Unmarshal([]byte(dataStr), &ev) == nil && ev.Delta != "" {
							if currentThinking.Len() > 0 {
								currentThinking.WriteString("\n")
							}
							currentThinking.WriteString(ev.Delta)
						}
					}
				case protocol.EventToken:
					var ev struct {
						Delta string `json:"delta"`
					}
					if json.Unmarshal([]byte(dataStr), &ev) == nil && ev.Delta != "" {
						currentOutput.WriteString(ev.Delta)
					}
					tokens++
				}
			}
			eventType = ""
			data.Reset()
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimSpace(line[7:])
		} else if strings.HasPrefix(line, "data: ") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(line[6:])
		}
	}

	// Flush any remaining content
	if currentThinking.Len() > 0 {
		entries = append(entries, MessageEntry{Type: "thinking", Content: currentThinking.String()})
	}
	if currentOutput.Len() > 0 {
		entries = append(entries, MessageEntry{Type: "output", Content: currentOutput.String()})
	}

	return entries, tokens, scanner.Err()
}

func consumeNDJSON(r io.Reader) (entries []MessageEntry, tokens int, err error) {
	scanner := bufio.NewScanner(r)
	var currentThinking strings.Builder
	var currentOutput strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		var ev protocol.StreamEvent
		if json.Unmarshal(line, &ev) != nil {
			continue
		}
		switch ev.Type {
		case protocol.EventThinking:
			if ev.Delta != "" {
				if currentThinking.Len() > 0 {
					currentThinking.WriteString("\n")
				}
				currentThinking.WriteString(ev.Delta)
			}
		case protocol.EventToken:
			currentOutput.WriteString(ev.Delta)
			tokens++
		}
	}

	// Flush any remaining content
	if currentThinking.Len() > 0 {
		entries = append(entries, MessageEntry{Type: "thinking", Content: currentThinking.String()})
	}
	if currentOutput.Len() > 0 {
		entries = append(entries, MessageEntry{Type: "output", Content: currentOutput.String()})
	}

	return entries, tokens, scanner.Err()
}
