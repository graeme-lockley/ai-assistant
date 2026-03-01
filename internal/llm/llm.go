package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Client calls the Deepseek API (OpenAI-compatible).
type Client struct {
	client *openai.Client
	model  string
}

// NewClient creates a Deepseek client with the given API key, base URL, and model.
// The HTTP client disables compression so streaming responses are delivered incrementally
// instead of being buffered by gzip decompression.
func NewClient(apiKey, baseURL, model string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("llm: API key is required")
	}
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	if cfg.BaseURL != "" && cfg.BaseURL[len(cfg.BaseURL)-1] == '/' {
		cfg.BaseURL = cfg.BaseURL[:len(cfg.BaseURL)-1]
	}
	// Disable gzip so the stream is not buffered by decompression; tokens arrive as the API sends them.
	cfg.HTTPClient = &http.Client{
		Transport: &http.Transport{DisableCompression: true},
	}
	return &Client{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}, nil
}

// Message represents a single chat message (role + content, and optionally tool_calls or tool result).
type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall // for assistant messages that request tool execution
	ToolCallID string     // for tool result messages
}

// ToolCall represents a single tool invocation from the LLM.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// Completer is the interface for generating assistant replies from conversation history.
// *Client implements it; tests can use a mock.
type Completer interface {
	Complete(ctx context.Context, messages []Message) (string, error)
}

// StreamCompleter is the interface for streaming assistant replies. *Client implements it.
type StreamCompleter interface {
	CompleteStream(ctx context.Context, messages []Message, sendDelta func(delta string) error) error
}

// StreamCompleterWithTools extends StreamCompleter with tool support. *Client implements it.
type StreamCompleterWithTools interface {
	StreamCompleter
	CompleteStreamWithTools(ctx context.Context, messages []Message, sendDelta func(delta string) error) (toolCalls []ToolCall, err error)
}

func messagesToOpenAI(messages []Message) []openai.ChatCompletionMessage {
	openaiMsgs := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	// Prepend system message with current date, time, and timezone so the model can reason about "now".
	now := time.Now()
	zoneName, _ := now.Zone()
	dateTimeCtx := fmt.Sprintf("Current date and time: %s (%s)", now.Format(time.RFC3339), zoneName)
	openaiMsgs = append(openaiMsgs, openai.ChatCompletionMessage{
		Role:    "system",
		Content: dateTimeCtx,
	})
	for _, m := range messages {
		content := m.Content
		// Deepseek (and some APIs) require "content" to be present; go-openai uses omitempty so empty string is omitted.
		// Use a single space when content is empty for assistant-with-tool-calls or tool-result messages.
		if content == "" && (m.Role == "tool" || (m.Role == "assistant" && len(m.ToolCalls) > 0)) {
			content = " "
		}
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: content,
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			tc := make([]openai.ToolCall, 0, len(m.ToolCalls))
			for _, t := range m.ToolCalls {
				tc = append(tc, openai.ToolCall{
					ID: t.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      t.Name,
						Arguments: t.Arguments,
					},
				})
			}
			msg.ToolCalls = tc
		}
		openaiMsgs = append(openaiMsgs, msg)
	}
	return openaiMsgs
}

// ToolDefinitions returns the fixed set of tools for the LLM.
func ToolDefinitions() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "web_search",
				Description: "Run a web search and return snippets and links.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"query": {Type: jsonschema.String, Description: "Search query"},
					},
					Required: []string{"query"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "web_get",
				Description: "Fetch a URL and return the response body as text.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"url": {Type: jsonschema.String, Description: "URL to fetch"},
					},
					Required: []string{"url"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "exec_bash",
				Description: "Run a bash command. Current working directory is the configured root.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"command": {Type: jsonschema.String, Description: "Bash command to run"},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "Read a file's contents. Path is relative to the configured root.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {Type: jsonschema.String, Description: "Relative path to the file"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_dir",
				Description: "List directory entries (names and types). Path is relative to the configured root.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {Type: jsonschema.String, Description: "Relative path to the directory"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "write_file",
				Description: "Create or overwrite a file. Path is relative to the configured root.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path":    {Type: jsonschema.String, Description: "Relative path to the file"},
						"content": {Type: jsonschema.String, Description: "Content to write"},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "merge_file",
				Description: "Insert or replace a region in a file. Use strategy 'replace' with start/end (1-based line numbers) or 'markers' with begin/end_marker line content.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path":        {Type: jsonschema.String, Description: "Relative path to the file"},
						"content":     {Type: jsonschema.String, Description: "Content to insert or replace with"},
						"strategy":    {Type: jsonschema.String, Description: "Either 'replace' or 'markers'"},
						"start":       {Type: jsonschema.Integer, Description: "Start line (1-based) for replace"},
						"end":         {Type: jsonschema.Integer, Description: "End line (1-based, inclusive) for replace"},
						"begin":       {Type: jsonschema.String, Description: "Line marker for markers strategy"},
						"end_marker":  {Type: jsonschema.String, Description: "End line marker for markers strategy"},
					},
					Required: []string{"path", "content", "strategy"},
				},
			},
		},
	}
}

// Complete sends the conversation history to the LLM and returns the assistant reply.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messagesToOpenAI(messages),
	}
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("llm complete: empty choices")
	}
	return resp.Choices[0].Message.Content, nil
}

// CompleteStream streams the assistant reply by calling sendDelta for each content delta.
func (c *Client) CompleteStream(ctx context.Context, messages []Message, sendDelta func(delta string) error) error {
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messagesToOpenAI(messages),
		Stream:   true,
	}
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("llm stream: %w", err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("llm stream recv: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		// Stream reasoning_content first (e.g. deepseek-reasoner), then content, so the user sees output as it arrives.
		if delta.ReasoningContent != "" {
			if err := sendDelta(delta.ReasoningContent); err != nil {
				return err
			}
		}
		if delta.Content != "" {
			if err := sendDelta(delta.Content); err != nil {
				return err
			}
		}
	}
}

// CompleteStreamWithTools streams the assistant reply and returns any tool calls.
// If the model requests tool execution, toolCalls is non-nil and the caller should run tools,
// append assistant message (with content + tool_calls) and tool result messages, then call again.
func (c *Client) CompleteStreamWithTools(ctx context.Context, messages []Message, sendDelta func(delta string) error) (toolCalls []ToolCall, err error) {
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messagesToOpenAI(messages),
		Tools:    ToolDefinitions(),
		Stream:   true,
	}
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("llm stream: %w", err)
	}
	defer stream.Close()

	// Accumulate tool calls (streamed in chunks: Index, ID, Function.Name, Function.Arguments)
	var acc []*ToolCall
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("llm stream recv: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		choice := chunk.Choices[0]

		if delta.ReasoningContent != "" {
			if err := sendDelta(delta.ReasoningContent); err != nil {
				return nil, err
			}
		}
		if delta.Content != "" {
			if err := sendDelta(delta.Content); err != nil {
				return nil, err
			}
		}
		for _, t := range delta.ToolCalls {
			idx := 0
			if t.Index != nil {
				idx = *t.Index
			}
			for len(acc) <= idx {
				acc = append(acc, &ToolCall{})
			}
			if t.ID != "" {
				acc[idx].ID = t.ID
			}
			if t.Function.Name != "" {
				acc[idx].Name = t.Function.Name
			}
			if t.Function.Arguments != "" {
				acc[idx].Arguments += t.Function.Arguments
			}
		}
		if choice.FinishReason == openai.FinishReasonToolCalls {
			result := make([]ToolCall, 0, len(acc))
			for _, t := range acc {
				if t.Name != "" {
					result = append(result, *t)
				}
			}
			return result, nil
		}
		if choice.FinishReason == openai.FinishReasonStop || choice.FinishReason == openai.FinishReasonLength || choice.FinishReason == openai.FinishReasonContentFilter {
			return nil, nil
		}
	}
	// Stream ended without explicit finish_reason tool_calls; if we accumulated any, return them
	if len(acc) > 0 {
		result := make([]ToolCall, 0, len(acc))
		for _, t := range acc {
			if t.Name != "" {
				result = append(result, *t)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}
	return nil, nil
}
