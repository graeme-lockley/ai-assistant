package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/graemelockley/ai-assistant/internal/llm"
	"github.com/graemelockley/ai-assistant/internal/tools"
)

// Agent is a per-session personal agent with conversation history.
type Agent struct {
	llm        llm.StreamCompleter
	runner     tools.Runner // optional; when set, agent uses tools
	summarizer llm.Summarizer
	bootstrap  string       // workspace core (SOUL, AGENT, IDENTITY) for system prompt
	history    []llm.Message
}

// New creates an agent that uses the given LLM stream completer.
// If runner is non-nil, the agent will use tools when the LLM requests them.
// If summarizer is non-nil, context compression will summarize dropped turns instead of discarding them.
// bootstrap is optional workspace context (e.g. LoadBootstrap) prepended to the system prompt each turn.
func New(llmClient llm.StreamCompleter, runner tools.Runner, summarizer llm.Summarizer, bootstrap string) *Agent {
	return &Agent{
		llm:        llmClient,
		runner:     runner,
		summarizer: summarizer,
		bootstrap:  bootstrap,
		history:    nil,
	}
}

// RespondStream appends the user message to history, streams the assistant reply via sendThinking (reasoning)
// and sendChunk (main content), then appends the full reply to history. When a tool runner is set, runs
// requested tools and continues until the LLM returns a final reply. model is an optional override; if empty,
// the LLM uses its default. History is compressed to stay within context limits before each LLM call.
func (a *Agent) RespondStream(ctx context.Context, userMessage string, sendThinking, sendChunk func(delta string) error, model string) error {
	a.history = append(a.history, llm.Message{
		Role:    "user",
		Content: userMessage,
	})
	// Compress so we stay under the model's context limit.
	if a.summarizer != nil {
		compressed, err := llm.CompressMessagesWithSummarizer(ctx, a.history, llm.DefaultMaxContextTokens, a.summarizer)
		if err != nil {
			log.Printf("[context] summarization failed, falling back to drop: %v", err)
			a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
		} else {
			a.history = compressed
		}
	} else {
		a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
	}

	var fullReply strings.Builder
	sendDelta := func(delta string) error {
		if _, err := fullReply.WriteString(delta); err != nil {
			return err
		}
		return sendChunk(delta)
	}

	if a.runner != nil {
		if withTools, ok := a.llm.(llm.StreamCompleterWithTools); ok {
			if err := a.respondStreamWithTools(ctx, withTools, sendThinking, sendDelta, model); err != nil {
				return err
			}
			// Keep history bounded after turn (tool results may have been appended).
			if a.summarizer != nil {
				compressed, err := llm.CompressMessagesWithSummarizer(ctx, a.history, llm.DefaultMaxContextTokens, a.summarizer)
				if err != nil {
					log.Printf("[context] summarization failed, falling back to drop: %v", err)
					a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
				} else {
					a.history = compressed
				}
			} else {
				a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
			}
			return nil
		}
	}

	if err := a.llm.CompleteStream(ctx, a.history, sendThinking, sendDelta, model, a.bootstrap); err != nil {
		return fmt.Errorf("agent respond: %w", err)
	}
	a.history = append(a.history, llm.Message{
		Role:    "assistant",
		Content: fullReply.String(),
	})
	if a.summarizer != nil {
		compressed, err := llm.CompressMessagesWithSummarizer(ctx, a.history, llm.DefaultMaxContextTokens, a.summarizer)
		if err != nil {
			log.Printf("[context] summarization failed, falling back to drop: %v", err)
			a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
		} else {
			a.history = compressed
		}
	} else {
		a.history = llm.CompressMessages(a.history, llm.DefaultMaxContextTokens)
	}
	return nil
}

// respondStreamWithTools runs the tool loop: call LLM with tools, stream deltas; if tool calls returned, run them and repeat.
// Assistant messages are appended with ReasoningContent and Content from the result so deepseek-reasoner receives
// reasoning_content on the next request.
func (a *Agent) respondStreamWithTools(ctx context.Context, withTools llm.StreamCompleterWithTools, sendThinking, sendDelta func(delta string) error, model string) error {
	for {
		res, err := withTools.CompleteStreamWithTools(ctx, a.history, sendThinking, sendDelta, model, a.bootstrap)
		if err != nil {
			return fmt.Errorf("agent respond: %w", err)
		}
		if res == nil {
			return nil
		}
		assistantMsg := llm.Message{
			Role:             "assistant",
			Content:          res.Content,
			ReasoningContent: res.ReasoningContent,
			ToolCalls:        res.ToolCalls,
		}
		a.history = append(a.history, assistantMsg)
		if len(res.ToolCalls) == 0 {
			return nil
		}
		// Run each tool and append tool result messages (truncated to avoid context explosion).
		for _, tc := range res.ToolCalls {
			toolLogTrunc := 200
			log.Printf("[tool] call id=%s name=%s args=%s", tc.ID, tc.Name, truncate(tc.Arguments, toolLogTrunc))
			toolResult, runErr := a.runner.Run(ctx, tc.Name, tc.Arguments)
			if runErr != nil {
				toolResult = "error: " + runErr.Error()
			}
			log.Printf("[tool] result id=%s name=%s result=%s err=%v", tc.ID, tc.Name, truncate(toolResult, toolLogTrunc), runErr)
			a.history = append(a.history, llm.Message{
				Role:       "tool",
				Content:    llm.TruncateToolResult(toolResult),
				ToolCallID: tc.ID,
			})
		}
	}
}

// truncate returns s truncated to at most maxLen runes, with "..." appended if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}
