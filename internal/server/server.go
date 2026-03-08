package server

import (
	"context"
	"encoding/json"
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
	"github.com/graemelockley/ai-assistant/internal/workspace"
)

// availableModels returns all known model IDs for /models.
func availableModels() []string {
	models := config.KnownModels()
	result := make([]string, len(models))
	for i, m := range models {
		result[i] = m.Name
	}
	return result
}

// Run starts the HTTP server and blocks until shutdown.
func Run(ctx context.Context, cfg config.Server) error {
	if cfg.DeepseekAPIKey == "" && cfg.AnthropicAPIKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY or ANTHROPIC_API_KEY is required")
	}

	var deepseekClient *llm.Client
	var anthropicClient *llm.AnthropicClient
	var err error

	if cfg.DeepseekAPIKey != "" {
		deepseekClient, err = llm.NewClient(cfg.DeepseekAPIKey, cfg.DeepseekBaseURL, cfg.DeepseekModel)
		if err != nil {
			return fmt.Errorf("deepseek llm: %w", err)
		}
	}

	if cfg.AnthropicAPIKey != "" {
		anthropicClient, err = llm.NewAnthropicClient(cfg.AnthropicAPIKey, config.DefaultAnthropicModel)
		if err != nil {
			return fmt.Errorf("anthropic llm: %w", err)
		}
	}

	// Determine default model
	defaultModel := cfg.DeepseekModel
	if defaultModel == "" {
		defaultModel = config.DefaultDeepseekModel
	}

	// Create unified provider
	multiProvider := llm.NewMultiProvider(deepseekClient, anthropicClient, defaultModel)

	var rootDir string
	if cfg.RootDir != "" {
		rootDir, err = workspace.ResolveRoot(cfg.RootDir)
		if err != nil {
			return fmt.Errorf("workspace root: %w", err)
		}
	} else {
		rootDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("root dir: %w", err)
		}
	}
	if err := workspace.Ensure(rootDir); err != nil {
		return fmt.Errorf("workspace init: %w", err)
	}
	toolRunner, err := tools.NewRunner(rootDir, config.SearchConfig{
		Provider: cfg.SearchProvider,
	})
	if err != nil {
		return fmt.Errorf("tools: %w", err)
	}

	store := session.NewStore(multiProvider, toolRunner, llm.SummarizerFromCompleter(multiProvider), rootDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/models", handleModels(cfg))
	mux.HandleFunc("/model", handleModel(store, cfg))
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
			sessionID, ag = store.Create(cfg.DeepseekModel)
			w.Header().Set(protocol.HeaderSessionID, sessionID)
		}

		// Request body: parse by Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = protocol.ContentTypeJSON
		}
		message, requestedModel, err := protocol.ParseRequestBody(r.Body, contentType)
		if err != nil {
			if strings.Contains(err.Error(), "unsupported content type") {
				http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// If model is specified in request and this is a new session, set it
		if newSession && requestedModel != "" {
			store.SetModel(sessionID, requestedModel)
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

		model := store.GetModel(sessionID)
		if model == "" {
			model = cfg.DeepseekModel
		}

		if useSSE {
			w.Header().Set("Content-Type", protocol.ContentTypeSSE)
			sw := protocol.NewSSEWriter(w)
			if newSession {
				_ = sw.WriteEvent(protocol.EventSession, map[string]string{"session_id": sessionID})
				if flusher != nil {
					flusher.Flush()
				}
			}
			sendThinking := func(delta string) error {
				if err := sw.WriteEvent(protocol.EventThinking, map[string]string{"delta": delta}); err != nil {
					return err
				}
				if flusher != nil {
					flusher.Flush()
				}
				return nil
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
			if err := ag.RespondStream(r.Context(), message, sendThinking, sendChunk, model); err != nil {
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
		sendThinking := func(delta string) error {
			if err := nw.WriteLine(protocol.StreamEvent{Type: protocol.EventThinking, Delta: delta}); err != nil {
				return err
			}
			if flusher != nil {
				flusher.Flush()
			}
			return nil
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
		if err := ag.RespondStream(r.Context(), message, sendThinking, sendChunk, model); err != nil {
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

// handleModels returns the HTTP handler for GET /models (list available models). No session required.
func handleModels(cfg config.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(availableModels())
	}
}

// handleModel returns the HTTP handler for GET /model (query current) and POST /model (set). Requires X-Session-Id.
func handleModel(store *session.Store, cfg config.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get(protocol.HeaderSessionID)
		if sessionID == "" {
			http.Error(w, "session required", http.StatusUnauthorized)
			return
		}
		if store.Get(sessionID) == nil {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			model := store.GetModel(sessionID)
			if model == "" {
				model = cfg.DeepseekModel
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"model": model})
			return
		case http.MethodPost:
			var body struct {
				Model string `json:"model"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			model := strings.TrimSpace(body.Model)
			if model == "" {
				http.Error(w, "model is required", http.StatusBadRequest)
				return
			}
			valid := false
			for _, m := range availableModels() {
				if m == model {
					valid = true
					break
				}
			}
			if !valid {
				http.Error(w, "unknown model", http.StatusBadRequest)
				return
			}
			store.SetModel(sessionID, model)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"model": model})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
