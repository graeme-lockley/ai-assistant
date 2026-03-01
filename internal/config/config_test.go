package config

import (
	"os"
	"testing"
)

func TestServerFromEnv_Defaults(t *testing.T) {
	// Clear relevant env vars so we get defaults
	os.Unsetenv("AI_ASSISTANT_BIND")
	os.Unsetenv("DEEPSEEK_API_KEY")
	os.Unsetenv("DEEPSEEK_BASE_URL")
	os.Unsetenv("DEEPSEEK_MODEL")

	cfg := ServerFromEnv()

	if cfg.BindAddr != DefaultBindAddr {
		t.Errorf("BindAddr: got %q, want %q", cfg.BindAddr, DefaultBindAddr)
	}
	if cfg.DeepseekAPIKey != "" {
		t.Errorf("DeepseekAPIKey: got %q, want empty", cfg.DeepseekAPIKey)
	}
	if cfg.DeepseekBaseURL != DefaultDeepseekURL {
		t.Errorf("DeepseekBaseURL: got %q, want %q", cfg.DeepseekBaseURL, DefaultDeepseekURL)
	}
	if cfg.DeepseekModel != DefaultDeepseekModel {
		t.Errorf("DeepseekModel: got %q, want %q", cfg.DeepseekModel, DefaultDeepseekModel)
	}
}

func TestServerFromEnv_Overrides(t *testing.T) {
	t.Setenv("AI_ASSISTANT_BIND", ":9999")
	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	t.Setenv("DEEPSEEK_BASE_URL", "https://custom.example.com")
	t.Setenv("DEEPSEEK_MODEL", "custom-model")

	cfg := ServerFromEnv()

	if cfg.BindAddr != ":9999" {
		t.Errorf("BindAddr: got %q, want :9999", cfg.BindAddr)
	}
	if cfg.DeepseekAPIKey != "test-key" {
		t.Errorf("DeepseekAPIKey: got %q, want test-key", cfg.DeepseekAPIKey)
	}
	if cfg.DeepseekBaseURL != "https://custom.example.com" {
		t.Errorf("DeepseekBaseURL: got %q", cfg.DeepseekBaseURL)
	}
	if cfg.DeepseekModel != "custom-model" {
		t.Errorf("DeepseekModel: got %q, want custom-model", cfg.DeepseekModel)
	}
}

func TestREPLFromEnv_Default(t *testing.T) {
	os.Unsetenv("AI_ASSISTANT_SERVER_ADDR")

	cfg := REPLFromEnv()

	if cfg.ServerAddr != DefaultServerAddr {
		t.Errorf("ServerAddr: got %q, want %q", cfg.ServerAddr, DefaultServerAddr)
	}
}

func TestREPLFromEnv_Override(t *testing.T) {
	t.Setenv("AI_ASSISTANT_SERVER_ADDR", "192.168.1.1:9000")

	cfg := REPLFromEnv()

	if cfg.ServerAddr != "192.168.1.1:9000" {
		t.Errorf("ServerAddr: got %q, want 192.168.1.1:9000", cfg.ServerAddr)
	}
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		key    string
		val    string
		def    string
		want   string
		setEnv bool
	}{
		{"KEY", "value", "default", "value", true},
		{"KEY", "", "default", "default", true},
		{"KEY", "  ", "default", "default", true},
		{"KEY", "  x  ", "default", "x", true},
		{"UNSET_KEY", "", "default", "default", false},
	}
	for _, tt := range tests {
		if tt.setEnv {
			t.Setenv(tt.key, tt.val)
		} else {
			os.Unsetenv(tt.key)
		}
		got := envOrDefault(tt.key, tt.def)
		if got != tt.want {
			t.Errorf("envOrDefault(%q, %q): got %q, want %q", tt.key, tt.def, got, tt.want)
		}
	}
}
