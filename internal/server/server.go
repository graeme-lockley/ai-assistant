package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/graemelockley/ai-assistant/internal/agent"
	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/llm"
	"github.com/graemelockley/ai-assistant/internal/protocol"
	"github.com/graemelockley/ai-assistant/internal/session"
	"github.com/graemelockley/ai-assistant/internal/tools"
)

// Run starts the HTTP server and blocks until shutdown.
func Run(ctx context.Context, cfg config.Server) error {
	if cfg.DeepseekAPIKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY is required")
	}
	llmClient, err := llm.NewClient(cfg.DeepseekAPIKey, cfg.DeepseekBaseURL, cfg.DeepseekModel)
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}
	rootDir := cfg.RootDir
	if rootDir == "" {
		rootDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("root dir: %w", err)
		}
	}
	toolRunner, err := tools.NewRunner(rootDir)
	if err != nil {
		return fmt.Errorf("tools: %w", err)
	}
	store := session.NewStore(llmClient, toolRunner)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleChat(store, cfg))

	srv := &http.Server{
		Addr:    cfg.BindAddr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	log.Printf("server listening on %s", cfg.BindAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

// handleChat returns the HTTP handler for POST / (chat turn). All responses are streamed.
func handleChat(store *session.Store, cfg config.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Explicit session close: X-Session-Close: true with X-Session-Id
		if r.Header.Get(protocol.HeaderSessionClose) == "true" {
			sessionID := r.Header.Get(protocol.HeaderSessionID)
			if sessionID != "" {
				store.Close(sessionID, "explicit")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Session: lookup or create
		sessionID := r.Header.Get(protocol.HeaderSessionID)
		var ag *agent.Agent
		var newSession bool
		if sessionID != "" {
			ag = store.Get(sessionID)
			if ag == nil {
				http.Error(w, "invalid or expired session", http.StatusUnauthorized)
				return
			}
		} else {
			newSession = true
			sessionID, ag = store.Create()
			w.Header().Set(protocol.HeaderSessionID, sessionID)
		}

		// Request body: parse by Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = protocol.ContentTypeJSON
		}
		message, err := protocol.ParseRequestBody(r.Body, contentType)
		if err != nil {
			if strings.Contains(err.Error(), "unsupported content type") {
				http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Response format: from Accept (config default added later)
		accept := r.Header.Get("Accept")
		if accept == "" && cfg.DefaultResponseType != "" {
			accept = cfg.DefaultResponseType
		}
		if accept == "" {
			accept = protocol.AcceptHeaderSSE
		}
		useSSE := strings.Contains(accept, "event-stream")
		useNDJSON := strings.Contains(accept, "application/json") && !useSSE
		if !useSSE && !useNDJSON {
			useSSE = true
		}

		// Stream response
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, _ := w.(http.Flusher)

		if useSSE {
			w.Header().Set("Content-Type", protocol.ContentTypeSSE)
			sw := protocol.NewSSEWriter(w)
			if newSession {
				_ = sw.WriteEvent(protocol.EventSession, map[string]string{"session_id": sessionID})
				if flusher != nil {
					flusher.Flush()
				}
			}
			sendChunk := func(delta string) error {
				if err := sw.WriteEvent(protocol.EventToken, map[string]string{"delta": delta}); err != nil {
					return err
				}
				if flusher != nil {
					flusher.Flush()
				}
				return nil
			}
			if err := ag.RespondStream(r.Context(), message, sendChunk); err != nil {
				_ = sw.WriteEvent(protocol.EventError, map[string]string{"error": err.Error()})
				if flusher != nil {
					flusher.Flush()
				}
				return
			}
			_ = sw.WriteEvent(protocol.EventDone, nil)
			if flusher != nil {
				flusher.Flush()
			}
			return
		}

		// NDJSON
		w.Header().Set("Content-Type", protocol.ContentTypeJSON)
		nw := protocol.NewNDJSONWriter(w)
		if newSession {
			_ = nw.WriteLine(protocol.StreamEvent{Type: protocol.EventSession, SessionID: sessionID})
			if flusher != nil {
				flusher.Flush()
			}
		}
		sendChunk := func(delta string) error {
			if err := nw.WriteLine(protocol.StreamEvent{Type: protocol.EventToken, Delta: delta}); err != nil {
				return err
			}
			if flusher != nil {
				flusher.Flush()
			}
			return nil
		}
		if err := ag.RespondStream(r.Context(), message, sendChunk); err != nil {
			_ = nw.WriteLine(protocol.StreamEvent{Type: protocol.EventError, Error: err.Error()})
			if flusher != nil {
				flusher.Flush()
			}
			return
		}
		_ = nw.WriteLine(protocol.StreamEvent{Type: protocol.EventDone})
		if flusher != nil {
			flusher.Flush()
		}
	}
}
