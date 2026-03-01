package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRunner_emptyRoot_usesCwd(t *testing.T) {
	r, err := NewRunner("")
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}

func TestRunner_resolve_rejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	r := &runner{root: dir}
	_, err := r.resolve("../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path outside root")
	}
}

func TestRunner_readFile_and_pathResolution(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "sub", "file.txt")
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fpath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	r := &runner{root: dir}
	ctx := context.Background()
	out, err := r.readFile(ctx, `{"path":"sub/file.txt"}`)
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	if out != "hello" {
		t.Errorf("got %q, want hello", out)
	}
}

func TestRunner_readDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &runner{root: dir}
	ctx := context.Background()
	out, err := r.readDir(ctx, `{"path":"."}`)
	if err != nil {
		t.Fatalf("readDir: %v", err)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "sub/") {
		t.Errorf("readDir output should contain a.txt and sub/: got %q", out)
	}
}

func TestRunner_writeFile(t *testing.T) {
	dir := t.TempDir()
	r := &runner{root: dir}
	ctx := context.Background()
	out, err := r.writeFile(ctx, `{"path":"out.txt","content":"written"}`)
	if err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	if out != "wrote out.txt" {
		t.Errorf("got %q", out)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "out.txt"))
	if string(data) != "written" {
		t.Errorf("file content: got %q", data)
	}
}

func TestRunner_Run_unknownTool(t *testing.T) {
	r := &runner{root: t.TempDir()}
	_, err := r.Run(context.Background(), "no_such_tool", "{}")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
