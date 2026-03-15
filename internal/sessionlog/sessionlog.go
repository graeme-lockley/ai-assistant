package sessionlog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AppendTurn appends a completed turn to the session log file under logsDir.
// logsDir is the absolute path to the workspace logs directory. When logsDir is
// empty, AppendTurn is a no-op so session logging can be disabled by configuration.
//
// Log file naming: logsDir/YYYY-MM-DD-<session-id>.md (date in UTC).
// Each call appends a human-readable markdown fragment containing the user
// message and assistant reply, plus a comment header with session metadata.
func AppendTurn(logsDir, sessionID, userMsg, assistantReply string, turnIndex int, ts time.Time) error {
	if logsDir == "" {
		return nil
	}
	if sessionID == "" {
		return fmt.Errorf("session id is empty")
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	date := ts.UTC().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, sessionID)
	path := filepath.Join(logsDir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open session log: %w", err)
	}
	defer f.Close()

	header := fmt.Sprintf("<!-- session: %s | turn: %d | %s -->\n\n", sessionID, turnIndex, ts.UTC().Format(time.RFC3339))
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("write session header: %w", err)
	}

	if _, err := f.WriteString("## User\n\n"); err != nil {
		return fmt.Errorf("write user header: %w", err)
	}
	if _, err := f.WriteString(userMsg); err != nil {
		return fmt.Errorf("write user message: %w", err)
	}
	if _, err := f.WriteString("\n\n## Assistant\n\n"); err != nil {
		return fmt.Errorf("write assistant header: %w", err)
	}
	if _, err := f.WriteString(assistantReply); err != nil {
		return fmt.Errorf("write assistant reply: %w", err)
	}
	if _, err := f.WriteString("\n\n---\n\n"); err != nil {
		return fmt.Errorf("write turn separator: %w", err)
	}
	return nil
}

