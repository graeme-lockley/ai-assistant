package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultBindAddr      = ":8080"
	DefaultServerAddr    = "127.0.0.1:8080"
	DefaultDeepseekURL   = "https://api.deepseek.com"
	DefaultDeepseekModel = "deepseek-chat"
	DefaultHistoryMax    = 1000
)

// Server holds configuration for the server personality.
type Server struct {
	BindAddr            string
	DeepseekAPIKey      string
	DeepseekBaseURL     string
	DeepseekModel       string
	DefaultResponseType string // optional; e.g. "text/event-stream" or "application/json"
	RootDir             string // root directory for file tools and exec_bash cwd; empty = process working directory
}

// REPL holds configuration for the REPL client.
type REPL struct {
	ServerAddr          string // host:port, e.g. "127.0.0.1:8080"
	ServerURL           string // optional full URL, e.g. "http://127.0.0.1:8080"; if set overrides ServerAddr for HTTP
	DefaultRequestType  string // optional; e.g. "application/json" or "text/plain"
	DefaultResponseType string // optional; e.g. "text/event-stream" or "application/json"
	HistoryFile         string // path to repl history file; default: <UserConfigDir>/ai-assistant/repl_history
	HistoryMaxSize      int    // max history entries to keep; default 1000
}

// ServerFromEnv loads server config from environment variables.
func ServerFromEnv() Server {
	s := Server{
		BindAddr:            envOrDefault("AI_ASSISTANT_BIND", DefaultBindAddr),
		DeepseekAPIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		DeepseekBaseURL:     envOrDefault("DEEPSEEK_BASE_URL", DefaultDeepseekURL),
		DeepseekModel:       envOrDefault("DEEPSEEK_MODEL", DefaultDeepseekModel),
		DefaultResponseType: os.Getenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE"),
		RootDir:             os.Getenv("AI_ASSISTANT_ROOT_DIR"),
	}
	return s
}

// REPLFromEnv loads REPL config from environment variables.
func REPLFromEnv() REPL {
	historyFile := os.Getenv("AI_ASSISTANT_REPL_HISTORY_FILE")
	if historyFile == "" {
		if dir, err := os.UserConfigDir(); err == nil {
			historyFile = filepath.Join(dir, "ai-assistant", "repl_history")
		}
	}
	historyMax := DefaultHistoryMax
	if s := os.Getenv("AI_ASSISTANT_REPL_HISTORY_MAX"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			historyMax = n
		}
	}
	return REPL{
		ServerAddr:          envOrDefault("AI_ASSISTANT_SERVER_ADDR", DefaultServerAddr),
		ServerURL:           os.Getenv("AI_ASSISTANT_SERVER_URL"),
		DefaultRequestType:  os.Getenv("AI_ASSISTANT_DEFAULT_REQUEST_TYPE"),
		DefaultResponseType: os.Getenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE"),
		HistoryFile:         historyFile,
		HistoryMaxSize:      historyMax,
	}
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
