package workspace

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

//go:embed template/*
var templateFS embed.FS

// RequiredDirs is the list of directories that must exist under the workspace root.
// context/ subdirs are required by context-loader-spec §7.
var RequiredDirs = []string{
	"logs",
	"memory",
	"skills",
	"tools",
	"context",
	"context/indexes",
	"context/routing",
	"context/cache",
}

// ResolveRoot returns the absolute workspace path, expanding a leading ~ to the user's home directory.
func ResolveRoot(root string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("workspace root is empty")
	}
	if root == "~" || strings.HasPrefix(root, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand home: %w", err)
		}
		if root == "~" {
			root = home
		} else {
			root = filepath.Join(home, root[2:])
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("workspace root: %w", err)
	}
	return abs, nil
}

// Ensure creates the workspace root if it does not exist and populates it from the
// embedded template. If the root already exists, it creates any missing required
// directories and does not overwrite existing files.
func Ensure(root string) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("workspace root: %w", err)
	}
	root = abs

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return initFromTemplate(root)
		}
		return fmt.Errorf("workspace root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace root is not a directory: %s", root)
	}

	// Root exists: ensure required directories exist; do not overwrite files.
	return ensureDirs(root)
}

// initFromTemplate creates root and copies all template files and dirs.
// Only writes files that do not already exist.
func initFromTemplate(root string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create workspace root: %w", err)
	}
	return copyTemplate(root, "template")
}

// copyTemplate copies entries from the embedded template at prefix into root.
// prefix is the path inside the embed (e.g. "template"). For each file we write
// only if the destination does not exist.
func copyTemplate(root, prefix string) error {
	return fs.WalkDir(templateFS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(prefix, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(root, filepath.FromSlash(rel))
		if d.IsDir() {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dest, err)
			}
			return nil
		}
		if _, err := os.Stat(dest); err == nil {
			return nil // file exists, do not overwrite
		}
		data, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		return nil
	})
}

// ensureDirs creates any missing required directories under root.
func ensureDirs(root string) error {
	for _, dir := range RequiredDirs {
		p := filepath.Join(root, filepath.FromSlash(dir))
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", p, err)
		}
	}
	return nil
}

// CoreFiles is the Ring 1 core self order (workspace-design §8, §9): SOUL → AGENT → IDENTITY.
var CoreFiles = []string{"SOUL.md", "AGENT.md", "IDENTITY.md"}

// Ring2Files is the Ring 2 supporting context order (workspace-design §8): USER → MEMORY → TASKS.
var Ring2Files = []string{"USER.md", "MEMORY.md", "TASKS.md"}

// BootstrapOptions controls how the system prompt is built. Zero value means Ring 1 only, no caps.
type BootstrapOptions struct {
	IncludeRing2          bool // include USER.md, MEMORY.md, TASKS.md
	Ring2MaxTokens        int  // per-file token cap for Ring 2; used only when IncludeRing2 is true
	SystemPromptMaxTokens int  // hard cap for entire system prompt; 0 means no cap
}

// minimalPrompt is the workspace-design §9 minimal system prompt (priority order and rules).
const minimalPrompt = `You are an AI agent operating inside a structured workspace.

The workspace defines your identity, behaviour, memory, skills, and tools.

Follow the workspace files in this priority order (earlier has greater gravity in case of conflict):

1. SOUL.md
2. AGENT.md
3. IDENTITY.md
4. USER.md
5. MEMORY.md

Rules

SOUL.md defines your beliefs, tone, and values.
AGENT.md defines how you operate in your role.
IDENTITY.md defines who you are.
USER.md describes the user.
MEMORY.md contains distilled long-term knowledge.

Skills and tools may be loaded when required.

Never load raw logs (unless explicitly stated otherwise in the prompt).

Use structured reasoning when solving problems.

Update TASKS.md when necessary.

Remain consistent with your identity and memory.

---
`

// charsPerToken is a conservative estimate (~4 chars per token for English).
const charsPerToken = 4

// estimateTokens returns an approximate token count for the string (runes / charsPerToken).
func estimateTokens(s string) int {
	n := utf8.RuneCountInString(s)
	if n <= 0 {
		return 0
	}
	return (n + charsPerToken - 1) / charsPerToken
}

// truncateToTokens truncates s to at most maxTokens (by runes), appending suffix if truncated.
func truncateToTokens(s string, maxTokens int, suffix string) string {
	if maxTokens <= 0 {
		return ""
	}
	maxRunes := maxTokens * charsPerToken
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + suffix
}

// readFileWithCap reads a file under root and truncates content to maxTokens. Returns empty string on error or empty file.
func readFileWithCap(root, name string, maxTokens int) string {
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return ""
	}
	if maxTokens > 0 && estimateTokens(content) > maxTokens {
		content = truncateToTokens(content, maxTokens, "\n\n[... truncated]")
	}
	return content
}

// BuildSystemPrompt reads workspace core (and optionally Ring 2) files and returns the system prompt string.
// Ring 1 (SOUL, AGENT, IDENTITY) is always included. Ring 2 (USER, MEMORY, TASKS) is included when opts.IncludeRing2 is true,
// with each file capped at opts.Ring2MaxTokens. The entire prompt is truncated from the end to opts.SystemPromptMaxTokens if set.
func BuildSystemPrompt(root string, opts BootstrapOptions) string {
	var b strings.Builder
	b.WriteString(minimalPrompt)
	for _, name := range CoreFiles {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}
		b.WriteString("## ")
		b.WriteString(name)
		b.WriteString("\n\n")
		b.WriteString(content)
		b.WriteString("\n\n")
	}
	if opts.IncludeRing2 && opts.Ring2MaxTokens > 0 {
		for _, name := range Ring2Files {
			content := readFileWithCap(root, name, opts.Ring2MaxTokens)
			if content == "" {
				continue
			}
			b.WriteString("## ")
			b.WriteString(name)
			b.WriteString("\n\n")
			b.WriteString(content)
			b.WriteString("\n\n")
		}
	}
	out := strings.TrimSuffix(b.String(), "\n\n")
	if opts.SystemPromptMaxTokens > 0 && estimateTokens(out) > opts.SystemPromptMaxTokens {
		out = truncateToTokens(out, opts.SystemPromptMaxTokens, "\n\n[... system prompt truncated]")
	}
	return out
}

// LoadBootstrap reads the workspace core self (SOUL.md, AGENT.md, IDENTITY.md) from root and returns
// a system prompt string: minimal prompt (workspace-design §9) plus file contents in order. Missing files are skipped.
// Used to bootstrap the session so the agent has identity and role context.
// Equivalent to BuildSystemPrompt(root, BootstrapOptions{}).
func LoadBootstrap(root string) string {
	return BuildSystemPrompt(root, BootstrapOptions{})
}
