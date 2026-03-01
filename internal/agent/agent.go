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

// RespondStream appends the user message to history, streams the assistant reply via sendChunk,
// then appends the full reply to history. When a tool runner is set, runs requested tools
// and continues until the LLM returns a final reply.
func (a *Agent) RespondStream(ctx context.Context, userMessage string, sendChunk func(delta string) error) error {
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
			if err := a.respondStreamWithTools(ctx, withTools, sendDelta, &fullReply); err != nil {
				return err
			}
			a.history = append(a.history, llm.Message{
				Role:    "assistant",
				Content: fullReply.String(),
			})
			return nil
		}
	}

	if err := a.llm.CompleteStream(ctx, a.history, sendDelta); err != nil {
		return fmt.Errorf("agent respond: %w", err)
	}
	a.history = append(a.history, llm.Message{
		Role:    "assistant",
		Content: fullReply.String(),
	})
	return nil
}

// respondStreamWithTools runs the tool loop: call LLM with tools, stream deltas; if tool calls returned, run them and repeat.
func (a *Agent) respondStreamWithTools(ctx context.Context, withTools llm.StreamCompleterWithTools, sendDelta func(delta string) error, fullReply *strings.Builder) error {
	for {
		toolCalls, err := withTools.CompleteStreamWithTools(ctx, a.history, sendDelta)
		if err != nil {
			return fmt.Errorf("agent respond: %w", err)
		}
		if len(toolCalls) == 0 {
			return nil
		}
		// Append assistant message (content so far + tool_calls)
		assistantContent := fullReply.String()
		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   assistantContent,
			ToolCalls: toolCalls,
		}
		a.history = append(a.history, assistantMsg)
		// Run each tool and append tool result messages
		for _, tc := range toolCalls {
			toolLogTrunc := 200
			log.Printf("[tool] call id=%s name=%s args=%s", tc.ID, tc.Name, truncate(tc.Arguments, toolLogTrunc))
			result, runErr := a.runner.Run(ctx, tc.Name, tc.Arguments)
			if runErr != nil {
				result = "error: " + runErr.Error()
			}
			log.Printf("[tool] result id=%s name=%s result=%s err=%v", tc.ID, tc.Name, truncate(result, toolLogTrunc), runErr)
			a.history = append(a.history, llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
		// Reset for next round (we only want to accumulate the final reply text)
		fullReply.Reset()
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
