package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// AnthropicClient calls the Anthropic API (Claude).
type AnthropicClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
	baseURL    string
}

// NewAnthropicClient creates an Anthropic client with the given API key and model.
func NewAnthropicClient(apiKey, model string) (*AnthropicClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: API key is required")
	}
	return &AnthropicClient{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://api.anthropic.com/v1",
	}, nil
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type      string `json:"type,omitempty"`
	Text      string `json:"text,omitempty"`
	Name      string `json:"name,omitempty"`
	Input     string `json:"input,omitempty"`
	ToolUseID string `json:"tool_use_id,omitempty"`
}

type anthropicTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

func (c *AnthropicClient) CompleteStream(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string) error {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = c.model
	}

	reqBody := anthropicRequest{
		Model:     modelToUse,
		MaxTokens: 4096,
		Messages:  convertMessagesToAnthropic(messages),
		Stream:    true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read line: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "event: message_stop" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var event map[string]any
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if eventType, ok := event["type"].(string); ok {
				if eventType == "content_block_delta" {
					if delta, ok := event["delta"].(map[string]any); ok {
						if text, ok := delta["text"].(string); ok && text != "" {
							if sendDelta != nil {
								if err := sendDelta(text); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}
}

func (c *AnthropicClient) CompleteStreamWithTools(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string) (*StreamWithToolsResult, error) {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = c.model
	}

	reqBody := anthropicRequest{
		Model:     modelToUse,
		MaxTokens: 4096,
		Messages:  convertMessagesToAnthropic(messages),
		Stream:    true,
		Tools:     convertToolsToAnthropic(ToolDefinitions()),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var contentBuf strings.Builder
	var toolCalls []ToolCall

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read line: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var event map[string]any
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if eventType, ok := event["type"].(string); ok {
				switch eventType {
				case "content_block_delta":
					if delta, ok := event["delta"].(map[string]any); ok {
						if text, ok := delta["text"].(string); ok && text != "" {
							contentBuf.WriteString(text)
							if sendDelta != nil {
								if err := sendDelta(text); err != nil {
									return nil, err
								}
							}
						}
					}
				case "tool_use":
					if toolUse, ok := event["tool_use"].(map[string]any); ok {
						tc := ToolCall{}
						if id, ok := toolUse["id"].(string); ok {
							tc.ID = id
						}
						if name, ok := toolUse["name"].(string); ok {
							tc.Name = name
						}
						toolCalls = append(toolCalls, tc)
					}
				case "tool_use_input":
					if idx, ok := event["index"].(float64); ok {
						input, _ := event["input"].(map[string]any)
						inputJSON, _ := json.Marshal(input)
						if idx < float64(len(toolCalls)) {
							toolCalls[int(idx)].Arguments = string(inputJSON)
						}
					}
				case "message_stop":
					if len(toolCalls) > 0 {
						return &StreamWithToolsResult{
							ToolCalls: toolCalls,
							Content:   contentBuf.String(),
						}, nil
					}
					return &StreamWithToolsResult{
						Content: contentBuf.String(),
					}, nil
				}
			}
		}
	}

	if len(toolCalls) > 0 {
		return &StreamWithToolsResult{
			ToolCalls: toolCalls,
			Content:   contentBuf.String(),
		}, nil
	}
	return &StreamWithToolsResult{
		Content: contentBuf.String(),
	}, nil
}

func convertMessagesToAnthropic(messages []Message) []anthropicMessage {
	result := make([]anthropicMessage, 0, len(messages))

	for _, m := range messages {
		if m.Role == "system" {
			continue
		}

		role := "user"
		if m.Role == "assistant" {
			role = "assistant"
		}

		content := []anthropicContentBlock{{
			Type: "text",
			Text: m.Content,
		}}

		// Handle tool calls
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				content = append(content, anthropicContentBlock{
					Type: "tool_use",
					Name: tc.Name,
				})
			}
		}

		// Handle tool results
		if m.Role == "tool" {
			content = []anthropicContentBlock{{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Text:      m.Content,
			}}
		}

		result = append(result, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}

	return result
}

func convertToolsToAnthropic(tools []openai.Tool) []anthropicTool {
	result := make([]anthropicTool, len(tools))
	for i, t := range tools {
		result[i] = anthropicTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
	}
	return result
}
