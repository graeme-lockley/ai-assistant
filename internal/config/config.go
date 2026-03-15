package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultBindAddr               = ":8080"
	DefaultServerAddr             = "127.0.0.1:8080"
	DefaultDeepseekURL            = "https://api.deepseek.com"
	DefaultDeepseekModel          = "deepseek-chat"
	DefaultAnthropicModel         = "claude-sonnet-4-6"
	DefaultHistoryMax             = 1000
	DefaultSystemPromptMaxTokens  = 4096
	DefaultRing2MaxTokens         = 500
)

// ModelInfo contains information about a model.
type ModelInfo struct {
	Name     string
	Provider string
}

// KnownModels returns all known models with their provider info.
// Only includes model IDs that are currently accepted by the APIs (deprecated IDs removed to avoid 404s).
func KnownModels() []ModelInfo {
	return []ModelInfo{
		{Name: "deepseek-chat", Provider: "deepseek"},
		{Name: "deepseek-reasoner", Provider: "deepseek"},
		{Name: "claude-sonnet-4-6", Provider: "anthropic"},
		{Name: "claude-opus-4-6", Provider: "anthropic"},
		{Name: "claude-haiku-4-5-20251001", Provider: "anthropic"},
	}
}

// GetModelProvider returns the provider for a given model name.
func GetModelProvider(model string) string {
	for _, m := range KnownModels() {
		if m.Name == model {
			return m.Provider
		}
	}
	return ""
}

type SearchProvider string

const (
	SearchProviderDuckDuckGo SearchProvider = "duckduckgo"
)

type SearchConfig struct {
	Provider     SearchProvider
	TavilyAPIKey string // required for web_search; set from TAVILY_API_KEY
}

// BootstrapConfig controls workspace system prompt assembly (workspace-design §8, §9).
// Ring 1 (SOUL, AGENT, IDENTITY) is always included. Ring 2 (USER, MEMORY, TASKS) is included by default so the assistant knows the user; set IncludeRing2 false to disable.
type BootstrapConfig struct {
	IncludeRing2         bool // include USER.md, MEMORY.md, TASKS.md (default true so assistant knows the user)
	Ring2MaxTokens       int  // per-file token cap for Ring 2 files; used only when IncludeRing2 is true
	SystemPromptMaxTokens int // hard cap for entire system prompt; 0 means no cap
}

// Server holds configuration for the server personality.
// RootDir is the workspace root; file tools and workspace layout use this path.
type Server struct {
	BindAddr            string
	DeepseekAPIKey      string
	DeepseekBaseURL     string
	DeepseekModel       string
	AnthropicAPIKey     string
	DefaultResponseType string
	RootDir             string         // workspace root; set from AI_ASSISTANT_WORKSPACE, else AI_ASSISTANT_ROOT_DIR, else ~/.ai-assistant.workspace
	SearchProvider      SearchProvider
	TavilyAPIKey        string         // for web_search; set from TAVILY_API_KEY
	Bootstrap           BootstrapConfig // workspace system prompt (Ring 1 + optional Ring 2)
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

// Ask holds configuration for the ask command (single-shot client).
type Ask struct {
	ServerURL           string // optional full URL, e.g. "http://127.0.0.1:8080"; if set overrides ServerAddr for HTTP
	Model               string // optional model override; empty means use server default
	DefaultRequestType  string // optional; e.g. "application/json" or "text/plain"
	DefaultResponseType string // optional; e.g. "text/event-stream" or "application/json"
}

// ServerFromEnv loads server config from environment variables.
func ServerFromEnv() Server {
	searchProvider := SearchProvider(strings.ToLower(strings.TrimSpace(os.Getenv("AI_ASSISTANT_SEARCH_PROVIDER"))))
	if searchProvider != SearchProviderDuckDuckGo {
		searchProvider = SearchProviderDuckDuckGo
	}
	rootDir := os.Getenv("AI_ASSISTANT_WORKSPACE")
	if rootDir == "" {
		rootDir = os.Getenv("AI_ASSISTANT_ROOT_DIR")
	}
	if rootDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			rootDir = filepath.Join(home, ".ai-assistant.workspace")
		}
	}
	s := Server{
		BindAddr:            envOrDefault("AI_ASSISTANT_BIND", DefaultBindAddr),
		DeepseekAPIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		DeepseekBaseURL:     envOrDefault("DEEPSEEK_BASE_URL", DefaultDeepseekURL),
		DeepseekModel:       envOrDefault("DEEPSEEK_MODEL", DefaultDeepseekModel),
		AnthropicAPIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		DefaultResponseType: os.Getenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE"),
		RootDir:             strings.TrimSpace(rootDir),
		SearchProvider:      searchProvider,
		TavilyAPIKey:        os.Getenv("TAVILY_API_KEY"),
		Bootstrap:           bootstrapFromEnv(),
	}
	return s
}

// bootstrapFromEnv loads BootstrapConfig from environment variables.
// Ring 2 (USER, MEMORY, TASKS) is on by default so the assistant knows the user; set AI_ASSISTANT_BOOTSTRAP_RING2=false to disable.
func bootstrapFromEnv() BootstrapConfig {
	includeRing2 := true
	if s := strings.TrimSpace(strings.ToLower(os.Getenv("AI_ASSISTANT_BOOTSTRAP_RING2"))); s == "false" || s == "0" || s == "no" {
		includeRing2 = false
	}
	systemPromptMax := DefaultSystemPromptMaxTokens
	if s := os.Getenv("AI_ASSISTANT_SYSTEM_PROMPT_MAX_TOKENS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			systemPromptMax = n
		}
	}
	ring2Max := DefaultRing2MaxTokens
	if s := os.Getenv("AI_ASSISTANT_RING2_MAX_TOKENS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ring2Max = n
		}
	}
	return BootstrapConfig{
		IncludeRing2:          includeRing2,
		Ring2MaxTokens:        ring2Max,
		SystemPromptMaxTokens: systemPromptMax,
	}
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

// AskFromEnv loads Ask config from environment variables.
func AskFromEnv() Ask {
	return Ask{
		ServerURL:           os.Getenv("AI_ASSISTANT_SERVER_URL"),
		Model:               os.Getenv("AI_ASSISTANT_MODEL"),
		DefaultRequestType:  os.Getenv("AI_ASSISTANT_DEFAULT_REQUEST_TYPE"),
		DefaultResponseType: os.Getenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE"),
	}
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
