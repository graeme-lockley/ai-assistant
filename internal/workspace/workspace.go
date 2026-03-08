package workspace

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// LoadBootstrap reads the workspace core self (SOUL.md, AGENT.md, IDENTITY.md) from root and returns
// a system prompt string: minimal prompt (workspace-design §9) plus file contents in order. Missing files are skipped.
// Used to bootstrap the session so the agent has identity and role context.
func LoadBootstrap(root string) string {
	var b strings.Builder
	b.WriteString("You are an AI agent operating inside a structured workspace. The workspace defines your identity, behaviour, memory, skills, and tools. Follow the workspace files in this priority order (earlier has greater gravity in conflict): 1. SOUL.md 2. AGENT.md 3. IDENTITY.md 4. USER.md 5. MEMORY.md. SOUL = beliefs, tone, values. AGENT = how you operate. IDENTITY = who you are. USER = the user. MEMORY = distilled long-term knowledge. Use structured reasoning. Update TASKS.md when necessary. Remain consistent with your identity and memory.\n\n---\n\n")
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
	return strings.TrimSuffix(b.String(), "\n\n")
}
