package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEnsure_NonexistentRoot_CreatesAndPopulates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new-workspace")

	err := Ensure(root)
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	// Core files must exist
	coreFiles := []string{"AGENT.md", "IDENTITY.md", "SOUL.md", "USER.md", "MEMORY.md", "TASKS.md", "SKILLS.md", "TOOLS.md", "WORKSPACE.md"}
	for _, name := range coreFiles {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing core file %s: %v", name, err)
		}
	}

	// Required dirs must exist
	for _, dir := range RequiredDirs {
		p := filepath.Join(root, filepath.FromSlash(dir))
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("missing dir %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", p)
		}
	}
}

func TestEnsure_ExistingRoot_CreatesMissingDirsOnly(t *testing.T) {
	root := t.TempDir()
	// Remove a couple of dirs that might exist from other runs; create only logs and memory
	if err := os.MkdirAll(filepath.Join(root, "logs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "memory"), 0755); err != nil {
		t.Fatal(err)
	}
	// Do not create context/indexes, context/routing, context/cache, skills, tools, context

	err := Ensure(root)
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	for _, dir := range RequiredDirs {
		p := filepath.Join(root, filepath.FromSlash(dir))
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("missing dir %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", p)
		}
	}

	// Template files were not copied (root already existed)
	agentPath := filepath.Join(root, "AGENT.md")
	if _, err := os.Stat(agentPath); err == nil {
		t.Error("AGENT.md should not exist when root existed before Ensure (we only create missing dirs)")
	}
}

func TestEnsure_ExistingFile_NotOverwritten(t *testing.T) {
	root := t.TempDir()
	customContent := []byte("# AGENT\n\nCustom content here.\n")
	agentPath := filepath.Join(root, "AGENT.md")
	if err := os.WriteFile(agentPath, customContent, 0644); err != nil {
		t.Fatal(err)
	}

	err := Ensure(root)
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	got, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("read AGENT.md: %v", err)
	}
	if string(got) != string(customContent) {
		t.Errorf("AGENT.md was overwritten: got %q", got)
	}
}

func TestEnsure_Idempotent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "idempotent")

	err := Ensure(root)
	if err != nil {
		t.Fatalf("Ensure first: %v", err)
	}
	err = Ensure(root)
	if err != nil {
		t.Fatalf("Ensure second: %v", err)
	}

	// Should still have template content in AGENT.md (not corrupted)
	agentPath := filepath.Join(root, "AGENT.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("read AGENT.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("AGENT.md is empty after second Ensure")
	}
	if !bytesContains(data, []byte("# AGENT")) || !bytesContains(data, []byte("mission")) {
		t.Errorf("AGENT.md unexpected content: %q", data)
	}
}

func TestEnsure_RootIsFile_ReturnsError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(root, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Ensure(root)
	if err == nil {
		t.Fatal("expected error when root is a file")
	}
}

func TestResolveRoot_Empty_ReturnsError(t *testing.T) {
	_, err := ResolveRoot("")
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func TestResolveRoot_AbsolutePath_ReturnsAbs(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveRoot(dir)
	if err != nil {
		t.Fatalf("ResolveRoot: %v", err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Errorf("ResolveRoot: got %q, want %q", got, abs)
	}
}

func bytesContains(b, sub []byte) bool {
	for i := 0; i <= len(b)-len(sub); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return true
		}
	}
	return false
}

func TestResolveRoot_ExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("UserHomeDir failed")
	}
	want := filepath.Join(home, ".ai-assistant.workspace")
	got, err := ResolveRoot("~/.ai-assistant.workspace")
	if err != nil {
		t.Fatalf("ResolveRoot: %v", err)
	}
	if got != want {
		t.Errorf("ResolveRoot(~/.ai-assistant.workspace): got %q, want %q", got, want)
	}
}

func TestLoadBootstrap_MatchesBuildSystemPrompt_ZeroOpts(t *testing.T) {
	root := t.TempDir()
	for _, name := range CoreFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte("# "+name+"\n\ncontent"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	gotLoad := LoadBootstrap(root)
	gotBuild := BuildSystemPrompt(root, BootstrapOptions{})
	if gotLoad != gotBuild {
		t.Error("LoadBootstrap and BuildSystemPrompt(root, zero opts) should return the same string")
	}
}

func TestBuildSystemPrompt_Ring1Only_IncludesMinimalAndCoreFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "SOUL.md"), []byte("soul content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENT.md"), []byte("agent content"), 0644); err != nil {
		t.Fatal(err)
	}
	// IDENTITY missing

	out := BuildSystemPrompt(root, BootstrapOptions{})
	if !strings.Contains(out, minimalPrompt) {
		t.Error("output should contain minimal prompt")
	}
	if !strings.Contains(out, "## SOUL.md") || !strings.Contains(out, "soul content") {
		t.Error("output should contain SOUL.md section")
	}
	if !strings.Contains(out, "## AGENT.md") || !strings.Contains(out, "agent content") {
		t.Error("output should contain AGENT.md section")
	}
	if strings.Contains(out, "## IDENTITY.md") {
		t.Error("output should not contain IDENTITY.md when file is missing")
	}
}

func TestBuildSystemPrompt_IncludeRing2_AddsUserMemoryTasks(t *testing.T) {
	root := t.TempDir()
	for _, name := range CoreFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte("core"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "USER.md"), []byte("user content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "MEMORY.md"), []byte("memory content"), 0644); err != nil {
		t.Fatal(err)
	}
	// TASKS.md missing

	opts := BootstrapOptions{IncludeRing2: true, Ring2MaxTokens: 500}
	out := BuildSystemPrompt(root, opts)
	if !strings.Contains(out, "## USER.md") || !strings.Contains(out, "user content") {
		t.Error("output should contain USER.md when IncludeRing2 true")
	}
	if !strings.Contains(out, "## MEMORY.md") || !strings.Contains(out, "memory content") {
		t.Error("output should contain MEMORY.md when IncludeRing2 true")
	}
	if strings.Contains(out, "## TASKS.md") {
		t.Error("output should not contain TASKS.md when file is missing")
	}
}

func TestBuildSystemPrompt_Ring2_PerFileCapTruncates(t *testing.T) {
	root := t.TempDir()
	for _, name := range CoreFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// USER.md content longer than 10 tokens (~40 chars)
	longUser := string(make([]byte, 200)) // 200 chars >> 10 tokens
	if err := os.WriteFile(filepath.Join(root, "USER.md"), []byte(longUser), 0644); err != nil {
		t.Fatal(err)
	}

	opts := BootstrapOptions{IncludeRing2: true, Ring2MaxTokens: 10}
	out := BuildSystemPrompt(root, opts)
	if !strings.Contains(out, "## USER.md") {
		t.Error("output should contain USER.md section")
	}
	if !strings.Contains(out, "[... truncated]") {
		t.Error("USER.md should be truncated when over Ring2MaxTokens")
	}
}

func TestBuildSystemPrompt_SystemPromptMaxTokens_TruncatesFromEnd(t *testing.T) {
	root := t.TempDir()
	// Large core content so total prompt exceeds cap
	big := string(make([]byte, 8000)) // ~2000 tokens
	for _, name := range CoreFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte(big), 0644); err != nil {
			t.Fatal(err)
		}
	}

	opts := BootstrapOptions{SystemPromptMaxTokens: 100}
	out := BuildSystemPrompt(root, opts)
	estTokens := (utf8.RuneCountInString(out) + charsPerToken - 1) / charsPerToken
	if estTokens > 120 {
		t.Errorf("output should be truncated to ~SystemPromptMaxTokens; got est %d tokens", estTokens)
	}
	if !strings.Contains(out, "[... system prompt truncated]") {
		t.Error("output should end with truncation suffix when over cap")
	}
}

func TestBuildSystemPrompt_MissingRing2Files_Skipped(t *testing.T) {
	root := t.TempDir()
	for _, name := range CoreFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// No USER.md, MEMORY.md, TASKS.md

	opts := BootstrapOptions{IncludeRing2: true, Ring2MaxTokens: 500}
	out := BuildSystemPrompt(root, opts)
	if strings.Contains(out, "## USER.md") || strings.Contains(out, "## MEMORY.md") || strings.Contains(out, "## TASKS.md") {
		t.Error("missing Ring 2 files should be skipped; output should not contain their headers")
	}
}
