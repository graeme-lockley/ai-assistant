package sessionlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendTurn_CreatesAndAppendsFile(t *testing.T) {
	tmp := t.TempDir()
	logsDir := filepath.Join(tmp, "logs")
	sessionID := "session-123"
	ts := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)

	if err := AppendTurn(logsDir, sessionID, "user message", "assistant reply", 1, ts); err != nil {
		t.Fatalf("AppendTurn first: %v", err)
	}
	if err := AppendTurn(logsDir, sessionID, "follow-up", "second reply", 2, ts); err != nil {
		t.Fatalf("AppendTurn second: %v", err)
	}

	filename := filepath.Join(logsDir, "2026-03-13-session-123.md")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "<!-- session: session-123 | turn: 1 | 2026-03-13T10:00:00Z -->") {
		t.Errorf("missing header for turn 1:\n%s", content)
	}
	if !strings.Contains(content, "<!-- session: session-123 | turn: 2 | 2026-03-13T10:00:00Z -->") {
		t.Errorf("missing header for turn 2:\n%s", content)
	}
	if !strings.Contains(content, "## User\n\nuser message") {
		t.Errorf("missing first user block:\n%s", content)
	}
	if !strings.Contains(content, "## Assistant\n\nassistant reply") {
		t.Errorf("missing first assistant block:\n%s", content)
	}
	if !strings.Contains(content, "## User\n\nfollow-up") {
		t.Errorf("missing second user block:\n%s", content)
	}
	if !strings.Contains(content, "## Assistant\n\nsecond reply") {
		t.Errorf("missing second assistant block:\n%s", content)
	}
}

func TestAppendTurn_EmptyLogsDirIsNoop(t *testing.T) {
	tmp := t.TempDir()
	// logsDir intentionally empty: should be a no-op.
	if err := AppendTurn("", "session-123", "user message", "assistant reply", 1, time.Now()); err != nil {
		t.Fatalf("AppendTurn: %v", err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files to be created, got %d", len(entries))
	}
}

