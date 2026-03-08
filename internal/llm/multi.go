package llm

import (
	"context"
	"strings"
)

// MultiProvider can switch between different LLM providers based on model name.
type MultiProvider struct {
	deepseek     *Client
	anthropic    *AnthropicClient
	defaultModel string
}

// NewMultiProvider creates a provider that can use both Deepseek and Anthropic.
func NewMultiProvider(deepseek *Client, anthropic *AnthropicClient, defaultModel string) *MultiProvider {
	return &MultiProvider{
		deepseek:     deepseek,
		anthropic:    anthropic,
		defaultModel: defaultModel,
	}
}

// Complete is required for Completer interface (used by SummarizerFromCompleter)
func (m *MultiProvider) Complete(ctx context.Context, messages []Message) (string, error) {
	provider := getProviderForModel(m.defaultModel)
	if provider == "anthropic" && m.anthropic != nil {
		var result strings.Builder
		err := m.anthropic.CompleteStream(ctx, messages, nil, func(delta string) error {
			result.WriteString(delta)
			return nil
		}, "")
		return result.String(), err
	}
	if m.deepseek != nil {
		return m.deepseek.Complete(ctx, messages)
	}
	return "", nil
}

// CompleteStream streams a response from the appropriate provider.
func (m *MultiProvider) CompleteStream(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string) error {
	if model == "" {
		model = m.defaultModel
	}
	provider := getProviderForModel(model)
	if provider == "anthropic" && m.anthropic != nil {
		return m.anthropic.CompleteStream(ctx, messages, sendThinking, sendDelta, model)
	}
	if m.deepseek != nil {
		return m.deepseek.CompleteStream(ctx, messages, sendThinking, sendDelta, model)
	}
	return nil
}

// CompleteStreamWithTools streams a response with tool support from the appropriate provider.
func (m *MultiProvider) CompleteStreamWithTools(ctx context.Context, messages []Message, sendThinking, sendDelta func(delta string) error, model string) (*StreamWithToolsResult, error) {
	if model == "" {
		model = m.defaultModel
	}
	provider := getProviderForModel(model)
	if provider == "anthropic" && m.anthropic != nil {
		return m.anthropic.CompleteStreamWithTools(ctx, messages, sendThinking, sendDelta, model)
	}
	if m.deepseek != nil {
		return m.deepseek.CompleteStreamWithTools(ctx, messages, sendThinking, sendDelta, model)
	}
	return nil, nil
}

func getProviderForModel(model string) string {
	// Default: if model starts with "claude", use anthropic
	if len(model) >= 6 && model[:6] == "claude" {
		return "anthropic"
	}
	// Default to deepseek
	return "deepseek"
}

// Ensure MultiProvider implements the required interfaces
var _ Completer = (*MultiProvider)(nil)
var _ StreamCompleter = (*MultiProvider)(nil)
var _ StreamCompleterWithTools = (*MultiProvider)(nil)
