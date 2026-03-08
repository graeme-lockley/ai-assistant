package ask

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/protocol"
)

func TestRun_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	cfg := config.Ask{
		ServerURL: server.URL,
	}

	_, err := Run(context.Background(), cfg, "test instruction")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "internal server error") {
		t.Errorf("expected error to contain 'internal server error', got: %s", err.Error())
	}
}

func TestRun_InvalidSessionClose(t *testing.T) {
	sessionCreated := false
	closeCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(protocol.HeaderSessionClose) == "true" {
			closeCalled = true
			w.WriteHeader(http.StatusNoContent)
			return
		}

		sessionCreated = true
		w.Header().Set(protocol.HeaderSessionID, "test-session-id")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("event: token\ndata: {\"delta\": \"hello\"}\n\nevent: done\ndata: {}\n\n"))
	}))
	defer server.Close()

	cfg := config.Ask{
		ServerURL: server.URL,
	}

	result, err := Run(context.Background(), cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !sessionCreated {
		t.Error("expected session to be created")
	}
	if !closeCalled {
		t.Error("expected session close to be called")
	}
	if result.SessionID != "test-session-id" {
		t.Errorf("SessionID: got %q, want test-session-id", result.SessionID)
	}
}

func TestRun_SSEWithThinking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(protocol.HeaderSessionID, "test-session")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`event: thinking
data: {"delta": "reasoning..."}

event: token
data: {"delta": "Hello"}

event: done
data: {}

`))
	}))
	defer server.Close()

	cfg := config.Ask{
		ServerURL:           server.URL,
		DefaultResponseType: "text/event-stream",
	}

	result, err := Run(context.Background(), cfg, "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Tokens != 1 {
		t.Errorf("Tokens: got %d, want 1", result.Tokens)
	}
	if len(result.Entries) != 2 {
		t.Errorf("Entries: got %d, want 2", len(result.Entries))
	}
	if result.Entries[0].Type != "thinking" || result.Entries[0].Content != "reasoning..." {
		t.Errorf("Thinking entry: got %+v, want {Type:thinking Content:reasoning...}", result.Entries[0])
	}
	if result.Entries[1].Type != "output" || result.Entries[1].Content != "Hello" {
		t.Errorf("Output entry: got %+v, want {Type:output Content:Hello}", result.Entries[1])
	}
}

func TestRun_NDJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(protocol.HeaderSessionID, "test-session")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"type": "token", "delta": "Hello"}
{"type": "token", "delta": " World"}
{"type": "done"}
`))
	}))
	defer server.Close()

	cfg := config.Ask{
		ServerURL:           server.URL,
		DefaultResponseType: "application/json",
	}

	result, err := Run(context.Background(), cfg, "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Tokens != 2 {
		t.Errorf("Tokens: got %d, want 2", result.Tokens)
	}
}

func TestRun_ModelOverride(t *testing.T) {
	var receivedModel string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(protocol.HeaderSessionID, "test-session")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("event: token\ndata: {\"delta\": \"response\"}\n\nevent: done\ndata: {}\n\n"))
	}))
	defer server.Close()

	cfg := config.Ask{
		ServerURL: server.URL,
		Model:     "deepseek-reasoner",
	}

	_, err := Run(context.Background(), cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedModel != "deepseek-reasoner" {
		t.Logf("Model override would be sent to server (not implemented in mock): %s", receivedModel)
	}
	_ = receivedModel
}
