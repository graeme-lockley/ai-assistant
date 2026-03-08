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
	"github.com/graemelockley/ai-assistant/internal/workspace"
)

// sessionEntry holds the agent and metadata for one session.
type sessionEntry struct {
	agent     *agent.Agent
	createdAt time.Time
	model     string // optional per-session model override; empty means use server default
}

// Store holds session ID -> agent mapping. Safe for concurrent use.
// rootDir is the workspace root; used to load SOUL/AGENT/IDENTITY into the system prompt for new sessions.
// LogOutput, when non-nil, is used for session lifecycle console messages instead of os.Stderr (for tests).
type Store struct {
	mu         sync.RWMutex
	agents     map[string]*sessionEntry
	llm        llm.StreamCompleter
	runner     tools.Runner
	summarizer llm.Summarizer
	rootDir    string
	logOutput  io.Writer
}

// NewStore creates a session store that creates agents using the given LLM stream completer.
// rootDir is the workspace root; agents get workspace core (SOUL, AGENT, IDENTITY) in the system prompt when non-empty.
// If runner is non-nil, agents will have access to tools. If summarizer is non-nil, context compression will summarize dropped turns.
func NewStore(llmClient llm.StreamCompleter, runner tools.Runner, summarizer llm.Summarizer, rootDir string) *Store {
	return &Store{
		agents:     make(map[string]*sessionEntry),
		llm:        llmClient,
		runner:     runner,
		summarizer: summarizer,
		rootDir:    rootDir,
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
// If model is non-empty, it sets the initial model for the session. Bootstrap (SOUL, AGENT, IDENTITY) is loaded from rootDir when set.
func (s *Store) Create(model string) (sessionID string, ag *agent.Agent) {
	bootstrap := ""
	if s.rootDir != "" {
		bootstrap = workspace.LoadBootstrap(s.rootDir)
	}
	ag = agent.New(s.llm, s.runner, s.summarizer, bootstrap)
	sessionID = uuid.New().String()
	now := time.Now()
	s.mu.Lock()
	s.agents[sessionID] = &sessionEntry{agent: ag, createdAt: now, model: model}
	s.mu.Unlock()
	if model != "" {
		fmt.Fprintf(s.logOut(), "%s [session] created %s model=%s\n", now.UTC().Format(time.RFC3339), sessionID, model)
	} else {
		fmt.Fprintf(s.logOut(), "%s [session] created %s\n", now.UTC().Format(time.RFC3339), sessionID)
	}
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

// GetModel returns the session's model override, or empty string if not set (caller should use server default).
func (s *Store) GetModel(sessionID string) string {
	s.mu.RLock()
	ent := s.agents[sessionID]
	s.mu.RUnlock()
	if ent == nil {
		return ""
	}
	return ent.model
}

// SetModel sets the model override for the session. Logs to the server console when the model changes.
// No-op if session not found.
func (s *Store) SetModel(sessionID string, model string) {
	s.mu.Lock()
	ent := s.agents[sessionID]
	if ent == nil {
		s.mu.Unlock()
		return
	}
	oldModel := ent.model
	ent.model = model
	s.mu.Unlock()
	if model != oldModel {
		ts := time.Now().UTC().Format(time.RFC3339)
		fmt.Fprintf(s.logOut(), "%s [session] %s model=%s\n", ts, sessionID, model)
	}
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
