package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
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

// Message represents a single chat message (role + content).
type Message struct {
	Role    string
	Content string
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

// Complete sends the conversation history to the LLM and returns the assistant reply.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	openaiMsgs := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, m := range messages {
		openaiMsgs = append(openaiMsgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMsgs,
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
	openaiMsgs := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, m := range messages {
		openaiMsgs = append(openaiMsgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMsgs,
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
