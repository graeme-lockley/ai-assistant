package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/graemelockley/ai-assistant/internal/config"
)

// TavilySearchURL is the Tavily Search API endpoint (can be overridden for testing).
var TavilySearchURL = "https://api.tavily.com/search"

// Runner runs the fixed set of tools. All file paths are resolved relative to the root directory.
type Runner interface {
	Run(ctx context.Context, toolName string, argsJSON string) (result string, err error)
}

// NewRunner returns a Runner that uses rootDir for file operations and exec_bash cwd.
// If rootDir is empty, the process working directory is used.
func NewRunner(rootDir string, searchCfg config.SearchConfig) (Runner, error) {
	if rootDir == "" {
		d, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("tools root dir: %w", err)
		}
		rootDir = d
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("tools root dir: %w", err)
	}
	return &runner{root: abs, searchCfg: searchCfg}, nil
}

type runner struct {
	root      string
	searchCfg config.SearchConfig
}

func (r *runner) resolve(path string) (string, error) {
	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path must be relative: %s", path)
	}
	abs := filepath.Join(r.root, cleaned)
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(r.root, abs)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("path outside root: %s", path)
	}
	return abs, nil
}

func (r *runner) Run(ctx context.Context, toolName string, argsJSON string) (string, error) {
	switch toolName {
	case "web_search":
		return r.webSearch(ctx, argsJSON)
	case "web_get":
		return r.webGet(ctx, argsJSON)
	case "exec_bash":
		return r.execBash(ctx, argsJSON)
	case "read_file":
		return r.readFile(ctx, argsJSON)
	case "read_dir":
		return r.readDir(ctx, argsJSON)
	case "write_file":
		return r.writeFile(ctx, argsJSON)
	case "merge_file":
		return r.mergeFile(ctx, argsJSON)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (r *runner) webSearch(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("web_search args: %w", err)
	}
	if r.searchCfg.TavilyAPIKey == "" {
		return "", fmt.Errorf("web_search: TAVILY_API_KEY is required; set it in the environment")
	}
	return r.tavilySearch(ctx, args.Query)
}

func (r *runner) tavilySearch(ctx context.Context, query string) (string, error) {
	body := struct {
		Query        string `json:"query"`
		SearchDepth  string `json:"search_depth"`
		MaxResults   int    `json:"max_results"`
	}{Query: query, SearchDepth: "basic", MaxResults: 10}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("web_search: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TavilySearchURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.searchCfg.TavilyAPIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_search request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("web_search: %s %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var data struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", fmt.Errorf("web_search parse: %w", err)
	}
	var out strings.Builder
	if data.Answer != "" {
		out.WriteString(strings.TrimSpace(data.Answer))
		out.WriteString("\n\n")
	}
	for _, res := range data.Results {
		if res.Title != "" || res.Content != "" {
			if res.Title != "" {
				out.WriteString(res.Title)
				if res.URL != "" {
					out.WriteString(" — ")
					out.WriteString(res.URL)
				}
				out.WriteString("\n")
			}
			if res.Content != "" {
				out.WriteString(strings.TrimSpace(res.Content))
				out.WriteString("\n\n")
			}
		}
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		result = "No search results found."
	}
	return result, nil
}

func (r *runner) webGet(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("web_get args: %w", err)
	}
	if args.URL == "" {
		return "", fmt.Errorf("web_get: url is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, args.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("web_get: status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (r *runner) execBash(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("exec_bash args: %w", err)
	}
	if args.Command == "" {
		return "", fmt.Errorf("exec_bash: command is required")
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", args.Command)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("exec_bash: %w", err)
	}
	return string(out), nil
}

func (r *runner) readFile(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("read_file args: %w", err)
	}
	path, err := r.resolve(args.Path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read_file: %w", err)
	}
	return string(data), nil
}

func (r *runner) readDir(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("read_dir args: %w", err)
	}
	path, err := r.resolve(args.Path)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("read_dir: %w", err)
	}
	var out strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			out.WriteString(name)
			out.WriteString("/\n")
		} else {
			out.WriteString(name)
			out.WriteString("\n")
		}
	}
	return strings.TrimSuffix(out.String(), "\n"), nil
}

func (r *runner) writeFile(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("write_file args: %w", err)
	}
	path, err := r.resolve(args.Path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("write_file mkdir: %w", err)
	}
	if err := os.WriteFile(path, []byte(args.Content), 0644); err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}
	return fmt.Sprintf("wrote %s", args.Path), nil
}

func (r *runner) mergeFile(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path      string `json:"path"`
		Content   string `json:"content"`
		Strategy  string `json:"strategy"`
		Start     int    `json:"start"`
		End       int    `json:"end"`
		Begin     string `json:"begin"`
		EndMarker string `json:"end_marker"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("merge_file args: %w", err)
	}
	path, err := r.resolve(args.Path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("merge_file read: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	switch args.Strategy {
	case "replace":
		if args.Start < 1 || args.End < args.Start {
			return "", fmt.Errorf("merge_file: invalid start/end (1-based)")
		}
		startIdx := args.Start - 1
		endIdx := args.End
		if endIdx > len(lines) {
			endIdx = len(lines)
		}
		newLines = append(newLines, lines[:startIdx]...)
		newLines = append(newLines, strings.Split(args.Content, "\n")...)
		newLines = append(newLines, lines[endIdx:]...)
	case "markers":
		if args.Begin == "" || args.EndMarker == "" {
			return "", fmt.Errorf("merge_file: begin and end_marker required for markers strategy")
		}
		var startIdx, endIdx int = -1, -1
		for i, l := range lines {
			if strings.TrimSpace(l) == strings.TrimSpace(args.Begin) {
				startIdx = i
				break
			}
		}
		for i := startIdx + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == strings.TrimSpace(args.EndMarker) {
				endIdx = i
				break
			}
		}
		if startIdx < 0 || endIdx < 0 {
			return "", fmt.Errorf("merge_file: could not find begin or end_marker")
		}
		newLines = append(newLines, lines[:startIdx+1]...)
		newLines = append(newLines, strings.Split(args.Content, "\n")...)
		newLines = append(newLines, lines[endIdx:]...)
	default:
		return "", fmt.Errorf("merge_file: strategy must be replace or markers")
	}
	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return "", fmt.Errorf("merge_file write: %w", err)
	}
	return fmt.Sprintf("merged %s", args.Path), nil
}
