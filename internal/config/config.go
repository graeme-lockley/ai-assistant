package config

import (
	"os"
	"strings"
)

const (
	DefaultBindAddr     = ":8080"
	DefaultServerAddr   = "127.0.0.1:8080"
	DefaultDeepseekURL  = "https://api.deepseek.com"
	DefaultDeepseekModel = "deepseek-chat"
)

// Server holds configuration for the server personality.
type Server struct {
	BindAddr            string
	DeepseekAPIKey      string
	DeepseekBaseURL     string
	DeepseekModel       string
	DefaultResponseType string // optional; e.g. "text/event-stream" or "application/json"
}

// REPL holds configuration for the REPL client.
type REPL struct {
	ServerAddr          string // host:port, e.g. "127.0.0.1:8080"
	ServerURL           string // optional full URL, e.g. "http://127.0.0.1:8080"; if set overrides ServerAddr for HTTP
	DefaultRequestType  string // optional; e.g. "application/json" or "text/plain"
	DefaultResponseType string // optional; e.g. "text/event-stream" or "application/json"
}

// ServerFromEnv loads server config from environment variables.
func ServerFromEnv() Server {
	s := Server{
		BindAddr:            envOrDefault("AI_ASSISTANT_BIND", DefaultBindAddr),
		DeepseekAPIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		DeepseekBaseURL:     envOrDefault("DEEPSEEK_BASE_URL", DefaultDeepseekURL),
		DeepseekModel:       envOrDefault("DEEPSEEK_MODEL", DefaultDeepseekModel),
		DefaultResponseType: os.Getenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE"),
	}
	return s
}

// REPLFromEnv loads REPL config from environment variables.
func REPLFromEnv() REPL {
	return REPL{
		ServerAddr:          envOrDefault("AI_ASSISTANT_SERVER_ADDR", DefaultServerAddr),
		ServerURL:           os.Getenv("AI_ASSISTANT_SERVER_URL"),
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
