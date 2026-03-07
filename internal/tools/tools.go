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

// Runner runs the fixed set of tools. All file paths are resolved relative to the root directory.
type Runner interface {
	Run(ctx context.Context, toolName string, argsJSON string) (result string, err error)
}

// NewRunner returns a Runner that uses rootDir for file operations and exec_bash cwd.
// If rootDir is empty, the process working directory is used.
// searchCfg provides search provider configuration.
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

// resolve resolves path relative to r.root and returns the absolute path if it is under root.
// Returns error if the result is outside root (e.g. path contains ".." escape).
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

	provider := r.searchCfg.Provider
	hasKey := r.hasSearchKey(provider)

	if provider == config.SearchProviderSerper && hasKey {
		return r.serperSearch(ctx, args.Query)
	}
	if provider == config.SearchProviderTavily && hasKey {
		return r.tavilySearch(ctx, args.Query)
	}
	if provider == config.SearchProviderGoogle && hasKey {
		return r.googleSearch(ctx, args.Query)
	}

	return r.duckDuckGoSearch(ctx, args.Query)
}

func (r *runner) hasSearchKey(provider config.SearchProvider) bool {
	switch provider {
	case config.SearchProviderSerper:
		return r.searchCfg.SerperAPIKey != ""
	case config.SearchProviderTavily:
		return r.searchCfg.TavilyAPIKey != ""
	case config.SearchProviderGoogle:
		return r.searchCfg.GoogleAPIKey != "" && r.searchCfg.GoogleCSEID != ""
	}
	return false
}

func (r *runner) duckDuckGoSearch(ctx context.Context, query string) (string, error) {
	url := "https://api.duckduckgo.com/?q=" + strings.ReplaceAll(query, " ", "+") + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_search request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data struct {
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
		Results []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("web_search parse: %w", err)
	}
	var out strings.Builder
	if data.AbstractText != "" {
		out.WriteString(data.AbstractText)
		if data.AbstractURL != "" {
			out.WriteString("\n")
			out.WriteString(data.AbstractURL)
		}
		out.WriteString("\n\n")
	}
	for i, t := range data.RelatedTopics {
		if i >= 10 {
			break
		}
		if t.Text != "" {
			out.WriteString(t.Text)
			if t.FirstURL != "" {
				out.WriteString(" ")
				out.WriteString(t.FirstURL)
			}
			out.WriteString("\n")
		}
	}
	for i, res := range data.Results {
		if i >= 10 {
			break
		}
		if res.Text != "" {
			out.WriteString(res.Text)
			if res.FirstURL != "" {
				out.WriteString(" ")
				out.WriteString(res.FirstURL)
			}
			out.WriteString("\n")
		}
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		result = "No instant answer or related results found for this query. The DuckDuckGo Instant Answer API has limited coverage (e.g. definitions, Wikipedia). For current news or broader web results, try rephrasing with more specific terms, or ask to fetch a specific article URL using the web_get tool."
	}
	return result, nil
}

func (r *runner) serperSearch(ctx context.Context, query string) (string, error) {
	url := "https://google.serper.dev/search"
	body := fmt.Sprintf(`{"q": "%s", "num": 10}`, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-Key", r.searchCfg.SerperAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("serper request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data struct {
		Organic []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Link    string `json:"link"`
		} `json:"organic"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", fmt.Errorf("serper parse: %w", err)
	}
	var out strings.Builder
	for i, res := range data.Organic {
		if i >= 10 {
			break
		}
		out.WriteString(res.Title)
		out.WriteString("\n")
		out.WriteString(res.Snippet)
		out.WriteString("\n")
		out.WriteString(res.Link)
		out.WriteString("\n\n")
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "No search results found.", nil
	}
	return result, nil
}

func (r *runner) tavilySearch(ctx context.Context, query string) (string, error) {
	url := "https://api.tavily.com/search"
	body := fmt.Sprintf(`{"query": "%s", "max_results": 10}`, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("API-Key", r.searchCfg.TavilyAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tavily request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data struct {
		Results []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			URL     string `json:"url"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", fmt.Errorf("tavily parse: %w", err)
	}
	var out strings.Builder
	for i, res := range data.Results {
		if i >= 10 {
			break
		}
		out.WriteString(res.Title)
		out.WriteString("\n")
		out.WriteString(res.Content)
		out.WriteString("\n")
		out.WriteString(res.URL)
		out.WriteString("\n\n")
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "No search results found.", nil
	}
	return result, nil
}

func (r *runner) googleSearch(ctx context.Context, query string) (string, error) {
	url := "https://customsearch.googleapis.com/customsearch/v1"
	url += "?key=" + r.searchCfg.GoogleAPIKey
	url += "&cx=" + r.searchCfg.GoogleCSEID
	url += "&q=" + strings.ReplaceAll(query, " ", "+")
	url += "&num=10"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google search request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data struct {
		Items []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Link    string `json:"link"`
		} `json:"items"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", fmt.Errorf("google search parse: %w", err)
	}
	var out strings.Builder
	for i, res := range data.Items {
		if i >= 10 {
			break
		}
		out.WriteString(res.Title)
		out.WriteString("\n")
		out.WriteString(res.Snippet)
		out.WriteString("\n")
		out.WriteString(res.Link)
		out.WriteString("\n\n")
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "No search results found.", nil
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
	// Return as text; assume UTF-8
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
		Strategy  string `json:"strategy"`   // "replace" (use start/end) or "markers" (use begin/end)
		Start     int    `json:"start"`      // 1-based line for replace
		End       int    `json:"end"`        // 1-based line for replace (inclusive)
		Begin     string `json:"begin"`      // line marker for markers strategy
		EndMarker string `json:"end_marker"` // line marker for markers strategy
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
