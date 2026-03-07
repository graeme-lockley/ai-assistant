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

func TestDuckDuckGoSearch_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"AbstractText":"test abstract","AbstractURL":"https://example.com","RelatedTopics":[],"Results":[]}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	result, err := r.duckDuckGoSearch(context.Background(), "test query")
	if err != nil {
		t.Fatalf("duckDuckGoSearch: %v", err)
	}
	if !strings.Contains(result, "test abstract") {
		t.Errorf("expected result to contain 'test abstract', got: %s", result)
	}
}

func TestDuckDuckGoSearch_emptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"AbstractText":"","AbstractURL":"","RelatedTopics":[],"Results":[]}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	result, err := r.duckDuckGoSearch(context.Background(), "obscure query")
	if err != nil {
		t.Fatalf("duckDuckGoSearch: %v", err)
	}
	if result != "No search results found." {
		t.Errorf("expected 'No search results found.', got: %s", result)
	}
}

func TestDuckDuckGoSearch_relatedTopicsAndResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"AbstractText": "Abstract text",
			"AbstractURL": "https://abstract.example.com",
			"RelatedTopics": [
				{"Text": "Related topic 1", "FirstURL": "https://related1.example.com"},
				{"Text": "Related topic 2", "FirstURL": "https://related2.example.com"}
			],
			"Results": [
				{"Text": "Result 1 text", "FirstURL": "https://result1.example.com"}
			]
		}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	result, err := r.duckDuckGoSearch(context.Background(), "test")
	if err != nil {
		t.Fatalf("duckDuckGoSearch: %v", err)
	}

	if !strings.Contains(result, "Abstract text") {
		t.Errorf("expected abstract text, got: %s", result)
	}
	if !strings.Contains(result, "Related topic 1") {
		t.Errorf("expected related topic 1, got: %s", result)
	}
	if !strings.Contains(result, "Result 1 text") {
		t.Errorf("expected result 1, got: %s", result)
	}
}

func TestGoogleSearch_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[{"title":"Google Result","snippet":"Google snippet","link":"https://google.example.com"}]}`))
	}))
	defer server.Close()

	r := &runner{
		root: t.TempDir(),
		searchCfg: config.SearchConfig{
			Provider:     config.SearchProviderGoogle,
			GoogleAPIKey: "google-key",
			GoogleCSEID:  "cse-id",
		},
	}

	originalURL := GoogleSearchBaseURL
	GoogleSearchBaseURL = server.URL
	defer func() { GoogleSearchBaseURL = originalURL }()

	result, err := r.googleSearch(context.Background(), "test query")
	if err != nil {
		t.Fatalf("googleSearch: %v", err)
	}
	if !strings.Contains(result, "Google Result") {
		t.Errorf("expected result to contain 'Google Result', got: %s", result)
	}
}

func TestGoogleSearch_noResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	r := &runner{
		root: t.TempDir(),
		searchCfg: config.SearchConfig{
			Provider:     config.SearchProviderGoogle,
			GoogleAPIKey: "test-key",
			GoogleCSEID:  "test-cse",
		},
	}

	originalURL := GoogleSearchBaseURL
	GoogleSearchBaseURL = server.URL
	defer func() { GoogleSearchBaseURL = originalURL }()

	result, err := r.googleSearch(context.Background(), "obscure query")
	if err != nil {
		t.Fatalf("googleSearch: %v", err)
	}
	if result != "No search results found." {
		t.Errorf("expected 'No search results found.', got: %s", result)
	}
}

func TestGoogleSearch_resultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		items := make([]map[string]interface{}, 15)
		for i := 0; i < 15; i++ {
			items[i] = map[string]interface{}{
				"title":   "Result " + string(rune('A'+i)),
				"snippet": "Snippet " + string(rune('A'+i)),
				"link":    "https://example.com/" + string(rune('A'+i)),
			}
		}
		resp := map[string]interface{}{"items": items}
		w.Write([]byte(mustJSON(resp)))
	}))
	defer server.Close()

	r := &runner{
		root: t.TempDir(),
		searchCfg: config.SearchConfig{
			Provider:     config.SearchProviderGoogle,
			GoogleAPIKey: "test-key",
			GoogleCSEID:  "test-cse",
		},
	}

	originalURL := GoogleSearchBaseURL
	GoogleSearchBaseURL = server.URL
	defer func() { GoogleSearchBaseURL = originalURL }()

	result, err := r.googleSearch(context.Background(), "test")
	if err != nil {
		t.Fatalf("googleSearch: %v", err)
	}

	for i := 0; i < 10; i++ {
		if !strings.Contains(result, "Result "+string(rune('A'+i))) {
			t.Errorf("expected result %d, got: %s", i, result)
		}
	}
	for i := 10; i < 15; i++ {
		if strings.Contains(result, "Result "+string(rune('A'+i))) {
			t.Errorf("did not expect result %d in output", i)
		}
	}
}

func TestHasGoogleKey(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SearchConfig
		expected bool
	}{
		{"no key, no cse", config.SearchConfig{Provider: config.SearchProviderGoogle, GoogleAPIKey: "", GoogleCSEID: ""}, false},
		{"only key", config.SearchConfig{Provider: config.SearchProviderGoogle, GoogleAPIKey: "key", GoogleCSEID: ""}, false},
		{"only cse", config.SearchConfig{Provider: config.SearchProviderGoogle, GoogleAPIKey: "", GoogleCSEID: "cse"}, false},
		{"both key and cse", config.SearchConfig{Provider: config.SearchProviderGoogle, GoogleAPIKey: "key", GoogleCSEID: "cse"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runner{root: t.TempDir(), searchCfg: tt.cfg}
			result := r.hasGoogleKey()
			if result != tt.expected {
				t.Errorf("hasGoogleKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWebSearch_usesDuckDuckGoByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"AbstractText":"DDG result","AbstractURL":"","RelatedTopics":[],"Results":[]}`))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalDDG := DuckDuckGoBaseURL
	originalGoogle := GoogleSearchBaseURL
	DuckDuckGoBaseURL = server.URL
	GoogleSearchBaseURL = server.URL
	defer func() {
		DuckDuckGoBaseURL = originalDDG
		GoogleSearchBaseURL = originalGoogle
	}()

	result, err := r.webSearch(context.Background(), `{"query":"test"}`)
	if err != nil {
		t.Fatalf("webSearch: %v", err)
	}
	if !strings.Contains(result, "DDG result") {
		t.Errorf("expected DuckDuckGo result, got: %s", result)
	}
}

func TestWebSearch_usesGoogleWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[{"title":"Google result","snippet":"snippet","link":"https://google.com"}]}`))
	}))
	defer server.Close()

	r := &runner{
		root: t.TempDir(),
		searchCfg: config.SearchConfig{
			Provider:     config.SearchProviderGoogle,
			GoogleAPIKey: "test-key",
			GoogleCSEID:  "test-cse",
		},
	}

	originalDDG := DuckDuckGoBaseURL
	originalGoogle := GoogleSearchBaseURL
	DuckDuckGoBaseURL = server.URL
	GoogleSearchBaseURL = server.URL
	defer func() {
		DuckDuckGoBaseURL = originalDDG
		GoogleSearchBaseURL = originalGoogle
	}()

	result, err := r.webSearch(context.Background(), `{"query":"test"}`)
	if err != nil {
		t.Fatalf("webSearch: %v", err)
	}
	if !strings.Contains(result, "Google result") {
		t.Errorf("expected Google result, got: %s", result)
	}
}

func TestWebSearch_fallsBackToDuckDuckGoWithoutGoogleKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"AbstractText":"fallback DDG","AbstractURL":"","RelatedTopics":[],"Results":[]}`))
	}))
	defer server.Close()

	r := &runner{
		root: t.TempDir(),
		searchCfg: config.SearchConfig{
			Provider:     config.SearchProviderGoogle,
			GoogleAPIKey: "",
			GoogleCSEID:  "",
		},
	}

	originalDDG := DuckDuckGoBaseURL
	originalGoogle := GoogleSearchBaseURL
	DuckDuckGoBaseURL = server.URL
	GoogleSearchBaseURL = server.URL
	defer func() {
		DuckDuckGoBaseURL = originalDDG
		GoogleSearchBaseURL = originalGoogle
	}()

	result, err := r.webSearch(context.Background(), `{"query":"test"}`)
	if err != nil {
		t.Fatalf("webSearch: %v", err)
	}
	if !strings.Contains(result, "fallback DDG") {
		t.Errorf("expected fallback to DuckDuckGo, got: %s", result)
	}
}

func TestWebSearch_invalidArgsJSON(t *testing.T) {
	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	_, err := r.webSearch(context.Background(), "{invalid json}")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "web_search args") {
		t.Errorf("expected 'web_search args' error, got: %v", err)
	}
}

func TestWebSearch_httpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	_, err := r.duckDuckGoSearch(context.Background(), "test")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestWebSearch_contextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.duckDuckGoSearch(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestWebSearch_malformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json}"))
	}))
	defer server.Close()

	r := &runner{
		root:      t.TempDir(),
		searchCfg: config.SearchConfig{},
	}

	originalURL := DuckDuckGoBaseURL
	DuckDuckGoBaseURL = server.URL
	defer func() { DuckDuckGoBaseURL = originalURL }()

	_, err := r.duckDuckGoSearch(context.Background(), "test")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
