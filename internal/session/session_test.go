package session

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/graemelockley/ai-assistant/internal/llm"
)

// mockStreamCompleter implements llm.StreamCompleter for tests (no real LLM calls).
type mockStreamCompleter struct{}

func (m *mockStreamCompleter) CompleteStream(ctx context.Context, messages []llm.Message, sendThinking, sendDelta func(delta string) error, model string) error {
	return nil
}

func TestCreate_logsTimestampAndSessionID(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	var buf bytes.Buffer
	store.SetLogOutput(&buf)

	sessionID, _ := store.Create()

	if sessionID == "" {
		t.Fatal("expected non-empty session ID")
	}
	out := buf.String()
	// RFC3339 timestamp at start, then " [session] created <id>"
	rx := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z \[session\] created ` + regexp.QuoteMeta(sessionID) + `\n$`)
	if !rx.MatchString(out) {
		t.Errorf("log line unexpected format: %q", out)
	}
}

func TestClose_logsTimestampSessionIDAndReason(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	var buf bytes.Buffer
	store.SetLogOutput(&buf)

	sessionID, _ := store.Create()
	buf.Reset()

	store.Close(sessionID, "explicit")

	out := buf.String()
	if !strings.Contains(out, "[session] closed") {
		t.Errorf("log line missing '[session] closed': %q", out)
	}
	if !strings.Contains(out, sessionID) {
		t.Errorf("log line missing session ID %q: %q", sessionID, out)
	}
	if !strings.Contains(out, "explicit") {
		t.Errorf("log line missing reason 'explicit': %q", out)
	}
	// RFC3339 timestamp at start
	if ok, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, out); !ok {
		t.Errorf("log line should start with RFC3339 timestamp: %q", out)
	}
}

func TestClose_emptyReason_logsWithoutReasonSuffix(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	var buf bytes.Buffer
	store.SetLogOutput(&buf)

	sessionID, _ := store.Create()
	buf.Reset()

	store.Close(sessionID, "")

	out := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(out), sessionID) {
		t.Errorf("log line should end with session ID (no reason): %q", out)
	}
	if strings.Count(out, sessionID) != 1 {
		t.Errorf("session ID should appear once: %q", out)
	}
}

func TestClose_unknownSession_noLogLine(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	var buf bytes.Buffer
	store.SetLogOutput(&buf)

	store.Close("nonexistent-session-id", "explicit")

	if buf.Len() != 0 {
		t.Errorf("Close with unknown session should not log; got %q", buf.String())
	}
}

func TestClose_removesSession(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	store.SetLogOutput(&bytes.Buffer{})

	sessionID, ag := store.Create()
	if store.Get(sessionID) != ag {
		t.Fatal("Get should return agent before Close")
	}

	store.Close(sessionID, "explicit")

	if store.Get(sessionID) != nil {
		t.Error("Get should return nil after Close")
	}
}

func TestGetModel_SetModel(t *testing.T) {
	store := NewStore(&mockStreamCompleter{}, nil, nil)
	store.SetLogOutput(&bytes.Buffer{})

	sessionID, _ := store.Create()
	if got := store.GetModel(sessionID); got != "" {
		t.Errorf("GetModel new session: got %q, want empty", got)
	}
	store.SetModel(sessionID, "deepseek-reasoner")
	if got := store.GetModel(sessionID); got != "deepseek-reasoner" {
		t.Errorf("GetModel after SetModel: got %q, want deepseek-reasoner", got)
	}
	store.SetModel(sessionID, "deepseek-chat")
	if got := store.GetModel(sessionID); got != "deepseek-chat" {
		t.Errorf("GetModel after second SetModel: got %q, want deepseek-chat", got)
	}
	if got := store.GetModel("nonexistent"); got != "" {
		t.Errorf("GetModel unknown session: got %q, want empty", got)
	}
}
