package session

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/graemelockley/ai-assistant/internal/agent"
	"github.com/graemelockley/ai-assistant/internal/llm"
	"github.com/graemelockley/ai-assistant/internal/tools"
)

// sessionEntry holds the agent and metadata for one session.
type sessionEntry struct {
	agent     *agent.Agent
	createdAt time.Time
}

// Store holds session ID -> agent mapping. Safe for concurrent use.
// LogOutput, when non-nil, is used for session lifecycle console messages instead of os.Stderr (for tests).
// If runner is non-nil, created agents can use tools (file ops, exec, web, etc.).
type Store struct {
	mu        sync.RWMutex
	agents    map[string]*sessionEntry
	llm       llm.StreamCompleter
	runner    tools.Runner
	logOutput io.Writer
}

// NewStore creates a session store that creates agents using the given LLM stream completer.
// If runner is non-nil, agents will have access to tools (web search, file ops, exec_bash, etc.).
func NewStore(llmClient llm.StreamCompleter, runner tools.Runner) *Store {
	return &Store{
		agents: make(map[string]*sessionEntry),
		llm:    llmClient,
		runner: runner,
	}
}

// SetLogOutput sets the writer for session lifecycle log lines. If nil, os.Stderr is used. Used by tests.
func (s *Store) SetLogOutput(w io.Writer) {
	s.logOutput = w
}

func (s *Store) logOut() io.Writer {
	if s.logOutput != nil {
		return s.logOutput
	}
	return os.Stderr
}

// Create creates a new session and returns its ID and agent. Caller must not use the ID for lookup until after Create returns.
// Logs to the server console with a timestamp when the session is created.
func (s *Store) Create() (sessionID string, ag *agent.Agent) {
	ag = agent.New(s.llm, s.runner)
	sessionID = uuid.New().String()
	now := time.Now()
	s.mu.Lock()
	s.agents[sessionID] = &sessionEntry{agent: ag, createdAt: now}
	s.mu.Unlock()
	fmt.Fprintf(s.logOut(), "%s [session] created %s\n", now.UTC().Format(time.RFC3339), sessionID)
	return sessionID, ag
}

// Get returns the agent for the session ID, or nil if not found.
func (s *Store) Get(sessionID string) *agent.Agent {
	s.mu.RLock()
	ent := s.agents[sessionID]
	s.mu.RUnlock()
	if ent == nil {
		return nil
	}
	return ent.agent
}

// Close removes the session and logs to the server console with a timestamp and optional reason.
// reason is logged as-is; use values like "explicit", "timeout", "disconnect". No-op if session not found.
func (s *Store) Close(sessionID string, reason string) {
	s.mu.Lock()
	_, ok := s.agents[sessionID]
	if ok {
		delete(s.agents, sessionID)
	}
	s.mu.Unlock()
	if !ok {
		return
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	out := s.logOut()
	if reason != "" {
		fmt.Fprintf(out, "%s [session] closed %s %s\n", ts, sessionID, reason)
	} else {
		fmt.Fprintf(out, "%s [session] closed %s\n", ts, sessionID)
	}
}