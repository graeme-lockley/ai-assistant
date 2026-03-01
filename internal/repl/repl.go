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
	"path/filepath"
	"strings"
	"syscall"

	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/protocol"
	"github.com/peterh/liner"
)

// Run connects to the server over HTTP and runs the read-send-receive-print loop until exit.
// Each turn is one POST request; the response is streamed (SSE or NDJSON). Session ID is
// sent on subsequent requests. Stdout is flushed after each token so the user sees
// streamed output immediately. Input uses a readline with history (Up/Down for history,
// Left/Right for line editing); history is persisted to cfg.HistoryFile and bounded by
// cfg.HistoryMaxSize.
func Run(ctx context.Context, cfg config.REPL) error {
	baseURL := cfg.ServerURL
	if baseURL == "" {
		baseURL = "http://" + cfg.ServerAddr
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	state := liner.NewLiner()
	defer state.Close()
	state.SetCtrlCAborts(false)

	// Load history from file (dedupe consecutive, keep last HistoryMaxSize)
	if cfg.HistoryFile != "" {
		loadHistory(state, cfg.HistoryFile, cfg.HistoryMaxSize)
	}

	client := &http.Client{}
	var sessionID string
	var lastHistoryLine string
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	fmt.Fprintln(os.Stderr, "Connected (HTTP). Enter a message and press Enter (Ctrl+C to exit).")

	for {
		select {
		case <-ctx.Done():
			saveHistory(state, cfg.HistoryFile, cfg.HistoryMaxSize)
			return nil
		default:
		}

		line, err := state.Prompt("> ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				continue
			}
			if err == io.EOF {
				saveHistory(state, cfg.HistoryFile, cfg.HistoryMaxSize)
				return nil
			}
			return fmt.Errorf("repl: read stdin: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Append to history only if not duplicate of previous (consecutive dedupe)
		if line != lastHistoryLine {
			state.AppendHistory(line)
			lastHistoryLine = line
		}
		saveHistory(state, cfg.HistoryFile, cfg.HistoryMaxSize)

		// Slash commands: /help (client-side), /exit, /models, /model (server)
		handled, exit := handleSlashCommand(ctx, baseURL, client, &sessionID, line, out)
		if exit {
			saveHistory(state, cfg.HistoryFile, cfg.HistoryMaxSize)
			return nil
		}
		if handled {
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

// slashCommandHelp is the client-side help text for /help.
const slashCommandHelp = `Slash commands:
  /help   Show this help.
  /exit   Close the session and exit the REPL.
  /models List available models.
  /model  Show the current model for this session.
  /model <name>  Set the model for this session.
`

// handleSlashCommand handles /exit, /models, /model, /help. Returns (true, true) to exit REPL,
// (true, false) if a slash command was handled and the loop should continue, (false, false) otherwise.
func handleSlashCommand(ctx context.Context, baseURL string, client *http.Client, sessionID *string, line string, out *bufio.Writer) (handled bool, exit bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "/") {
		return false, false
	}
	parts := strings.Fields(line)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(line[len(cmd):])
	}

	switch cmd {
	case "/help":
		fmt.Fprint(os.Stderr, slashCommandHelp)
		return true, false
	case "/exit":
		if *sessionID != "" {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/", nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "exit: %v\n", err)
				return true, true
			}
			req.Header.Set(protocol.HeaderSessionID, *sessionID)
			req.Header.Set(protocol.HeaderSessionClose, "true")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "exit: %v\n", err)
				return true, true
			}
			resp.Body.Close()
		}
		return true, true
	case "/models":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "models: %v\n", err)
			return true, false
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "models: %v\n", err)
			return true, false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "models: %s\n", string(body))
			return true, false
		}
		var models []string
		if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
			fmt.Fprintf(os.Stderr, "models: %v\n", err)
			return true, false
		}
		for _, m := range models {
			fmt.Fprintln(out, m)
		}
		_ = out.Flush()
		return true, false
	case "/model":
		if *sessionID == "" {
			fmt.Fprintln(os.Stderr, "No active session. Send a message first.")
			return true, false
		}
		if arg == "" {
			// GET current model
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/model", nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "model: %v\n", err)
				return true, false
			}
			req.Header.Set(protocol.HeaderSessionID, *sessionID)
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "model: %v\n", err)
				return true, false
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				fmt.Fprintf(os.Stderr, "model: %s\n", string(body))
				return true, false
			}
			var v struct {
				Model string `json:"model"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
				fmt.Fprintf(os.Stderr, "model: %v\n", err)
				return true, false
			}
			fmt.Fprintf(out, "Current model: %s\n", v.Model)
			_ = out.Flush()
			return true, false
		}
		// POST set model
		payload, _ := json.Marshal(map[string]string{"model": arg})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/model", bytes.NewReader(payload))
		if err != nil {
			fmt.Fprintf(os.Stderr, "model: %v\n", err)
			return true, false
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(protocol.HeaderSessionID, *sessionID)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "model: %v\n", err)
			return true, false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "model: %s\n", string(respBody))
			return true, false
		}
		var v struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			fmt.Fprintf(os.Stderr, "model: %v\n", err)
			return true, false
		}
		fmt.Fprintf(out, "Model set to %s\n", v.Model)
		_ = out.Flush()
		return true, false
	}
	return false, false
}

// consumeSSE reads Server-Sent Events from r and prints token deltas to out, flushing after each
// so streamed output appears immediately. Thinking/reasoning tokens are printed in light grey (ANSI).
// Calls onSession with session_id from session events.
func consumeSSE(r io.Reader, out *bufio.Writer, onSession func(sessionID string)) error {
	const (
		ansiLightGrey = "\033[90m"
		ansiReset     = "\033[0m"
	)
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
				case protocol.EventThinking:
					var v struct {
						Delta string `json:"delta"`
					}
					if json.Unmarshal([]byte(dataStr), &v) == nil {
						_, _ = out.WriteString(ansiLightGrey)
						_, _ = out.WriteString(v.Delta)
						_, _ = out.WriteString(ansiReset)
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
// so streamed output appears immediately. Thinking/reasoning tokens are printed in light grey (ANSI).
// Calls onSession with session_id from session events.
func consumeNDJSON(r io.Reader, out *bufio.Writer, onSession func(sessionID string)) error {
	const (
		ansiLightGrey = "\033[90m"
		ansiReset     = "\033[0m"
	)
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
		case protocol.EventThinking:
			_, _ = out.WriteString(ansiLightGrey)
			_, _ = out.WriteString(ev.Delta)
			_, _ = out.WriteString(ansiReset)
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

// loadHistory reads history from path, dedupes consecutive lines, keeps the last maxLines,
// and loads them into state. Ignores errors (e.g. file not found).
func loadHistory(state *liner.State, path string, maxLines int) {
	if path == "" || maxLines <= 0 {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var deduped []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(deduped) > 0 && deduped[len(deduped)-1] == line {
			continue
		}
		deduped = append(deduped, line)
	}
	if scanner.Err() != nil {
		return
	}
	start := 0
	if len(deduped) > maxLines {
		start = len(deduped) - maxLines
	}
	var buf bytes.Buffer
	for i := start; i < len(deduped); i++ {
		buf.WriteString(deduped[i])
		buf.WriteByte('\n')
	}
	if buf.Len() > 0 {
		state.ReadHistory(&buf)
	}
}

// saveHistory writes state's history to path and trims the file to the last maxLines entries.
// Creates the parent directory if needed. No-op if path is empty.
func saveHistory(state *liner.State, path string, maxLines int) {
	if path == "" {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	state.WriteHistory(f)
	if err := f.Close(); err != nil {
		return
	}
	trimHistoryFile(path, maxLines)
}

// trimHistoryFile keeps only the last maxLines lines in the file at path.
func trimHistoryFile(path string, maxLines int) {
	if maxLines <= 0 {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	if len(lines) <= maxLines {
		return
	}
	lines = lines[len(lines)-maxLines:]
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		return
	}
}
