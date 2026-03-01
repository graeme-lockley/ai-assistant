package repl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/protocol"
)

// Run connects to the server over HTTP and runs the read-send-receive-print loop until exit.
// Each turn is one POST request; the response is streamed (SSE or NDJSON). Session ID is
// sent on subsequent requests. Stdout is flushed after each token so the user sees
// streamed output immediately.
func Run(ctx context.Context, cfg config.REPL) error {
	baseURL := cfg.ServerURL
	if baseURL == "" {
		baseURL = "http://" + cfg.ServerAddr
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := &http.Client{}
	scanner := bufio.NewScanner(os.Stdin)
	var sessionID string
	// Flushable stdout so streamed tokens appear immediately (no wait for newline/buffer full).
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	fmt.Fprintln(os.Stderr, "Connected (HTTP). Enter a message and press Enter (Ctrl+C to exit).")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		fmt.Fprint(os.Stdout, "> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("repl: read stdin: %w", err)
			}
			return nil
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

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
			payload, _ := json.Marshal(&protocol.Request{Message: line})
			body = bytes.NewReader(payload)
		} else {
			body = strings.NewReader(line)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/", body)
		if err != nil {
			return fmt.Errorf("repl: build request: %w", err)
		}
		req.Header.Set("Accept", accept)
		req.Header.Set("Content-Type", contentType)
		if sessionID != "" {
			req.Header.Set(protocol.HeaderSessionID, sessionID)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("repl: request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Fprintln(os.Stderr, "Session expired or invalid; starting new session.")
			sessionID = ""
			continue
		}
		if resp.StatusCode != http.StatusOK {
			errBody, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: %s\n", string(errBody))
			continue
		}

		// Capture session ID from header (first response)
		if id := resp.Header.Get(protocol.HeaderSessionID); id != "" {
			sessionID = id
		}

		// Consume stream
		if strings.Contains(accept, "event-stream") {
			if err := consumeSSE(resp.Body, out, func(sessionIDFromEvent string) {
				if sessionIDFromEvent != "" {
					sessionID = sessionIDFromEvent
				}
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		} else {
			if err := consumeNDJSON(resp.Body, out, func(sessionIDFromEvent string) {
				if sessionIDFromEvent != "" {
					sessionID = sessionIDFromEvent
				}
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		}
		_ = out.Flush()
		fmt.Println()
	}
}

// consumeSSE reads Server-Sent Events from r and prints token deltas to out, flushing after each
// so streamed output appears immediately. Calls onSession with session_id from session events.
func consumeSSE(r io.Reader, out *bufio.Writer, onSession func(sessionID string)) error {
	scanner := bufio.NewScanner(r)
	var eventType string
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// End of event
			if eventType != "" {
				dataStr := strings.TrimSpace(data.String())
				switch eventType {
				case protocol.EventSession:
					var v struct {
						SessionID string `json:"session_id"`
					}
					if json.Unmarshal([]byte(dataStr), &v) == nil {
						onSession(v.SessionID)
					}
				case protocol.EventToken:
					var v struct {
						Delta string `json:"delta"`
					}
					if json.Unmarshal([]byte(dataStr), &v) == nil {
						_, _ = out.WriteString(v.Delta)
						_ = out.Flush()
					}
				case protocol.EventError:
					var v struct {
						Error string `json:"error"`
					}
					if json.Unmarshal([]byte(dataStr), &v) == nil && v.Error != "" {
						fmt.Fprintf(os.Stderr, "Error: %s\n", v.Error)
					}
				case protocol.EventDone:
					// no-op
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
	return scanner.Err()
}

// consumeNDJSON reads NDJSON lines from r and prints token deltas to out, flushing after each
// so streamed output appears immediately. Calls onSession with session_id from session events.
func consumeNDJSON(r io.Reader, out *bufio.Writer, onSession func(sessionID string)) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev protocol.StreamEvent
		if json.Unmarshal(line, &ev) != nil {
			continue
		}
		switch ev.Type {
		case protocol.EventSession:
			onSession(ev.SessionID)
		case protocol.EventToken:
			_, _ = out.WriteString(ev.Delta)
			_ = out.Flush()
		case protocol.EventError:
			if ev.Error != "" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ev.Error)
			}
		case protocol.EventDone:
			// no-op
		}
	}
	return scanner.Err()
}
