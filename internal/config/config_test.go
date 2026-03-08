package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerFromEnv_Defaults(t *testing.T) {
	// Clear relevant env vars so we get defaults
	os.Unsetenv("AI_ASSISTANT_BIND")
	os.Unsetenv("AI_ASSISTANT_WORKSPACE")
	os.Unsetenv("AI_ASSISTANT_ROOT_DIR")
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
	// Default workspace root when neither env is set (same logic as ServerFromEnv)
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("UserHomeDir failed, cannot verify default RootDir")
	}
	wantRoot := filepath.Join(home, ".ai-assistant.workspace")
	if cfg.RootDir != wantRoot {
		t.Errorf("RootDir: got %q, want %q", cfg.RootDir, wantRoot)
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

func TestServerFromEnv_RootDir_WorkspaceOverride(t *testing.T) {
	t.Setenv("AI_ASSISTANT_WORKSPACE", "/custom/workspace")
	os.Unsetenv("AI_ASSISTANT_ROOT_DIR")

	cfg := ServerFromEnv()

	if cfg.RootDir != "/custom/workspace" {
		t.Errorf("RootDir: got %q, want /custom/workspace", cfg.RootDir)
	}
}

func TestServerFromEnv_RootDir_RootDirFallback(t *testing.T) {
	os.Unsetenv("AI_ASSISTANT_WORKSPACE")
	t.Setenv("AI_ASSISTANT_ROOT_DIR", "/legacy/root")

	cfg := ServerFromEnv()

	if cfg.RootDir != "/legacy/root" {
		t.Errorf("RootDir: got %q, want /legacy/root", cfg.RootDir)
	}
}

func TestServerFromEnv_RootDir_WorkspaceTakesPrecedence(t *testing.T) {
	t.Setenv("AI_ASSISTANT_WORKSPACE", "/workspace")
	t.Setenv("AI_ASSISTANT_ROOT_DIR", "/root")

	cfg := ServerFromEnv()

	if cfg.RootDir != "/workspace" {
		t.Errorf("RootDir: got %q, want /workspace (AI_ASSISTANT_WORKSPACE takes precedence)", cfg.RootDir)
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

func TestAskFromEnv_Default(t *testing.T) {
	os.Unsetenv("AI_ASSISTANT_SERVER_URL")
	os.Unsetenv("AI_ASSISTANT_MODEL")
	os.Unsetenv("AI_ASSISTANT_DEFAULT_REQUEST_TYPE")
	os.Unsetenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE")

	cfg := AskFromEnv()

	if cfg.ServerURL != "" {
		t.Errorf("ServerURL: got %q, want empty", cfg.ServerURL)
	}
	if cfg.Model != "" {
		t.Errorf("Model: got %q, want empty", cfg.Model)
	}
	if cfg.DefaultRequestType != "" {
		t.Errorf("DefaultRequestType: got %q, want empty", cfg.DefaultRequestType)
	}
	if cfg.DefaultResponseType != "" {
		t.Errorf("DefaultResponseType: got %q, want empty", cfg.DefaultResponseType)
	}
}

func TestAskFromEnv_Override(t *testing.T) {
	t.Setenv("AI_ASSISTANT_SERVER_URL", "http://custom:8080")
	t.Setenv("AI_ASSISTANT_MODEL", "deepseek-reasoner")
	t.Setenv("AI_ASSISTANT_DEFAULT_REQUEST_TYPE", "text/plain")
	t.Setenv("AI_ASSISTANT_DEFAULT_RESPONSE_TYPE", "application/json")

	cfg := AskFromEnv()

	if cfg.ServerURL != "http://custom:8080" {
		t.Errorf("ServerURL: got %q, want http://custom:8080", cfg.ServerURL)
	}
	if cfg.Model != "deepseek-reasoner" {
		t.Errorf("Model: got %q, want deepseek-reasoner", cfg.Model)
	}
	if cfg.DefaultRequestType != "text/plain" {
		t.Errorf("DefaultRequestType: got %q, want text/plain", cfg.DefaultRequestType)
	}
	if cfg.DefaultResponseType != "application/json" {
		t.Errorf("DefaultResponseType: got %q, want application/json", cfg.DefaultResponseType)
	}
}
