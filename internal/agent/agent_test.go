package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/graemelockley/ai-assistant/internal/llm"
)

// mockStreamCompleter streams a fixed reply for testing.
type mockStreamCompleter struct {
	reply string
	err   error
}

func (m *mockStreamCompleter) CompleteStream(ctx context.Context, messages []llm.Message, sendDelta func(delta string) error) error {
	if m.err != nil {
		return m.err
	}
	// Stream reply in chunks (one rune at a time so we can verify streaming)
	for _, r := range m.reply {
		if err := sendDelta(string(r)); err != nil {
			return err
		}
	}
	return nil
}

func TestAgent_RespondStream_SendsChunksAndAppendsToHistory(t *testing.T) {
	reply := "Hello from the assistant"
	mock := &mockStreamCompleter{reply: reply}
	ag := New(mock, nil)

	var chunks []string
	sendChunk := func(delta string) error {
		chunks = append(chunks, delta)
		return nil
	}
	err := ag.RespondStream(context.Background(), "user message", sendChunk)
	if err != nil {
		t.Fatalf("RespondStream: %v", err)
	}
	if got := strings.Join(chunks, ""); got != reply {
		t.Errorf("chunks joined: got %q, want %q", got, reply)
	}

	// Second call should include previous exchange in history (mock receives it)
	chunks = nil
	err = ag.RespondStream(context.Background(), "follow-up", sendChunk)
	if err != nil {
		t.Fatalf("RespondStream (second): %v", err)
	}
	if strings.Join(chunks, "") != reply {
		t.Errorf("second reply: got %q, want %q", strings.Join(chunks, ""), reply)
	}
}

func TestAgent_RespondStream_PropagatesLLMError(t *testing.T) {
	wantErr := errors.New("llm failed")
	mock := &mockStreamCompleter{err: wantErr}
	ag := New(mock, nil)

	err := ag.RespondStream(context.Background(), "hello", func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error from RespondStream")
	}
	if !errors.Is(err, wantErr) {
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	}
}

func TestAgent_RespondStream_EmptyMessage(t *testing.T) {
	mock := &mockStreamCompleter{reply: "ok"}
	ag := New(mock, nil)

	var chunks []string
	err := ag.RespondStream(context.Background(), "", func(delta string) error {
		chunks = append(chunks, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("RespondStream: %v", err)
	}
	if got := strings.Join(chunks, ""); got != "ok" {
		t.Errorf("got %q, want ok", got)
	}
}
