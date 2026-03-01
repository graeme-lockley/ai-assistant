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
	llm    llm.StreamCompleter
	runner tools.Runner // optional; when set, agent uses tools
	history []llm.Message
}

// New creates an agent that uses the given LLM stream completer.
// If runner is non-nil, the agent will use tools (web search, file ops, exec, etc.) when the LLM requests them.
func New(llmClient llm.StreamCompleter, runner tools.Runner) *Agent {
	return &Agent{
		llm:    llmClient,
		runner: runner,
		history: nil,
	}
}

// RespondStream appends the user message to history, streams the assistant reply via sendThinking (reasoning)
// and sendChunk (main content), then appends the full reply to history. When a tool runner is set, runs
// requested tools and continues until the LLM returns a final reply. model is an optional override; if empty,
// the LLM uses its default.
func (a *Agent) RespondStream(ctx context.Context, userMessage string, sendThinking, sendChunk func(delta string) error, model string) error {
	a.history = append(a.history, llm.Message{
		Role:    "user",
		Content: userMessage,
	})
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
			return nil
		}
	}

	if err := a.llm.CompleteStream(ctx, a.history, sendThinking, sendDelta, model); err != nil {
		return fmt.Errorf("agent respond: %w", err)
	}
	a.history = append(a.history, llm.Message{
		Role:    "assistant",
		Content: fullReply.String(),
	})
	return nil
}

// respondStreamWithTools runs the tool loop: call LLM with tools, stream deltas; if tool calls returned, run them and repeat.
// Assistant messages are appended with ReasoningContent and Content from the result so deepseek-reasoner receives
// reasoning_content on the next request.
func (a *Agent) respondStreamWithTools(ctx context.Context, withTools llm.StreamCompleterWithTools, sendThinking, sendDelta func(delta string) error, model string) error {
	for {
		res, err := withTools.CompleteStreamWithTools(ctx, a.history, sendThinking, sendDelta, model)
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
		// Run each tool and append tool result messages
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
				Content:    toolResult,
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
