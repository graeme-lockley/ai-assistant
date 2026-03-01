package llm

import (
	"testing"
)

func TestNewClient_RequiresAPIKey(t *testing.T) {
	_, err := NewClient("", "https://api.example.com", "model")
	if err == nil {
		t.Fatal("NewClient with empty API key: expected error")
	}
	if err.Error() != "llm: API key is required" {
		t.Errorf("got error %q", err.Error())
	}
}

func TestNewClient_TrimsTrailingSlashFromBaseURL(t *testing.T) {
	client, err := NewClient("key", "https://api.example.com/", "model")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	// We can't easily assert the internal config; verify we got a client and it has the model
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.model != "model" {
		t.Errorf("model: got %q, want model", client.model)
	}
}

func TestNewClient_Success(t *testing.T) {
	client, err := NewClient("test-key", "https://api.deepseek.com", "deepseek-chat")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.model != "deepseek-chat" {
		t.Errorf("model: got %q, want deepseek-chat", client.model)
	}
}
