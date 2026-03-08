package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graemelockley/ai-assistant/internal/config"
)

func TestTavilySearch_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer token, got %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Query != "test query" {
			t.Errorf("expected query 'test query', got %q", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"title":"Test Result","url":"https://example.com","content":"Test content here."}],"answer":""}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "test-key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	result, err := r.tavilySearch(context.Background(), "test query")
	if err != nil {
		t.Fatalf("tavilySearch: %v", err)
	}
	if !strings.Contains(result, "Test Result") {
		t.Errorf("expected result to contain 'Test Result', got: %s", result)
	}
	if !strings.Contains(result, "Test content here.") {
		t.Errorf("expected result to contain content, got: %s", result)
	}
}

func TestTavilySearch_withAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[],"answer":"Leo Messi is a footballer."}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	result, err := r.tavilySearch(context.Background(), "who is Leo Messi")
	if err != nil {
		t.Fatalf("tavilySearch: %v", err)
	}
	if !strings.Contains(result, "Leo Messi is a footballer.") {
		t.Errorf("expected answer in result, got: %s", result)
	}
}

func TestTavilySearch_emptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[],"answer":""}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	result, err := r.tavilySearch(context.Background(), "obscure query")
	if err != nil {
		t.Fatalf("tavilySearch: %v", err)
	}
	if result != "No search results found." {
		t.Errorf("expected 'No search results found.', got: %s", result)
	}
}

func TestWebSearch_noAPIKey_returnsError(t *testing.T) {
	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{}, // no TavilyAPIKey
	}
	_, err := r.webSearch(context.Background(), `{"query":"test"}`)
	if err == nil {
		t.Fatal("expected error when TAVILY_API_KEY not set")
	}
	if !strings.Contains(err.Error(), "TAVILY_API_KEY") {
		t.Errorf("expected TAVILY_API_KEY in error, got: %v", err)
	}
}

func TestWebSearch_usesTavilyWhenKeySet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"title":"Tavily result","url":"https://tavily.com","content":"From Tavily."}],"answer":""}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "test-key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	result, err := r.webSearch(context.Background(), `{"query":"test"}`)
	if err != nil {
		t.Fatalf("webSearch: %v", err)
	}
	if !strings.Contains(result, "Tavily result") {
		t.Errorf("expected Tavily result, got: %s", result)
	}
}

func TestWebSearch_invalidArgsJSON(t *testing.T) {
	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	_, err := r.webSearch(context.Background(), "{invalid json}")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "web_search args") {
		t.Errorf("expected 'web_search args' error, got: %v", err)
	}
}

func TestTavilySearch_httpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	_, err := r.tavilySearch(context.Background(), "test")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestTavilySearch_contextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.tavilySearch(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestTavilySearch_malformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json}"))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{TavilyAPIKey: "key"},
	}
	originalURL := TavilySearchURL
	TavilySearchURL = server.URL
	defer func() { TavilySearchURL = originalURL }()

	_, err := r.tavilySearch(context.Background(), "test")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}
