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
	"time"

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
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type      string      `json:"type,omitempty"`
	Text      string      `json:"text"`                   // text block (required when type is text; do not omit when empty)
	Content   string      `json:"content,omitempty"`      // tool_result body (API expects "content", not "text")
	Name      string      `json:"name,omitempty"`
	ID        string      `json:"id,omitempty"`             // tool_use block id (required when type is tool_use)
	Input     interface{} `json:"input,omitempty"`           // tool_use input as object
	ToolUseID string      `json:"tool_use_id,omitempty"`    // tool_result block
}

type anthropicTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

func (c *AnthropicClient) CompleteStream(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string, systemPrompt string) error {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = c.model
	}

	systemContent := systemPrompt
	if systemContent != "" {
		now := time.Now()
		zoneName, _ := now.Zone()
		systemContent = fmt.Sprintf("Current date and time: %s (%s)", now.Format(time.RFC3339), zoneName) + "\n\n" + systemContent
	}
	reqBody := anthropicRequest{
		Model:     modelToUse,
		MaxTokens: 4096,
		System:    systemContent,
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

func (c *AnthropicClient) CompleteStreamWithTools(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string, systemPrompt string) (*StreamWithToolsResult, error) {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = c.model
	}

	systemContent := systemPrompt
	if systemContent != "" {
		now := time.Now()
		zoneName, _ := now.Zone()
		systemContent = fmt.Sprintf("Current date and time: %s (%s)", now.Format(time.RFC3339), zoneName) + "\n\n" + systemContent
	}
	reqBody := anthropicRequest{
		Model:     modelToUse,
		MaxTokens: 4096,
		System:    systemContent,
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
	// Anthropic streams tool_use via content_block_start (with content_block.type "tool_use") and
	// content_block_delta (with delta.type "input_json_delta", partial_json). Index is the content block index.
	var toolCallsByIndex = make(map[int]*ToolCall)
	var toolCallsOrder []int

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

			eventType, _ := event["type"].(string)
			switch eventType {
			case "content_block_start":
				if block, ok := event["content_block"].(map[string]any); ok {
					if blockType, _ := block["type"].(string); blockType == "tool_use" {
						idx, _ := event["index"].(float64)
						tc := &ToolCall{}
						if id, ok := block["id"].(string); ok {
							tc.ID = id
						}
						if name, ok := block["name"].(string); ok {
							tc.Name = name
						}
						i := int(idx)
						toolCallsByIndex[i] = tc
						toolCallsOrder = append(toolCallsOrder, i)
					}
				}
			case "content_block_delta":
				idx, _ := event["index"].(float64)
				if delta, ok := event["delta"].(map[string]any); ok {
					deltaType, _ := delta["type"].(string)
					switch deltaType {
					case "text_delta":
						if text, ok := delta["text"].(string); ok && text != "" {
							contentBuf.WriteString(text)
							if sendDelta != nil {
								if err := sendDelta(text); err != nil {
									return nil, err
								}
							}
						}
					case "input_json_delta":
						if partial, ok := delta["partial_json"].(string); ok {
							if tc := toolCallsByIndex[int(idx)]; tc != nil {
								tc.Arguments += partial
							}
						}
					}
				}
			case "message_stop":
				// Build result in content block order
				var toolCalls []ToolCall
				for _, i := range toolCallsOrder {
					if tc := toolCallsByIndex[i]; tc != nil {
						toolCalls = append(toolCalls, *tc)
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
		}
	}

	// EOF: build result if we have tool calls
	var toolCalls []ToolCall
	for _, i := range toolCallsOrder {
		if tc := toolCallsByIndex[i]; tc != nil {
			toolCalls = append(toolCalls, *tc)
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

		// Handle tool calls (assistant message with tool_use blocks: id, name, input required by API)
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				var inputObj interface{}
				if tc.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &inputObj)
				}
				content = append(content, anthropicContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: inputObj,
				})
			}
		}

		// Handle tool results (API expects "content" for the result body, not "text")
		if m.Role == "tool" {
			content = []anthropicContentBlock{{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
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
