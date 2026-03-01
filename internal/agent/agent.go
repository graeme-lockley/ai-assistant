package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/graemelockley/ai-assistant/internal/llm"
)

// Agent is a per-session personal agent with conversation history.
type Agent struct {
	llm     llm.StreamCompleter
	history []llm.Message
}

// New creates an agent that uses the given LLM stream completer.
func New(llmClient llm.StreamCompleter) *Agent {
	return &Agent{
		llm:     llmClient,
		history: nil,
	}
}

// RespondStream appends the user message to history, streams the assistant reply via sendChunk,
// then appends the full reply to history.
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
	if err := a.llm.CompleteStream(ctx, a.history, sendDelta); err != nil {
		return fmt.Errorf("agent respond: %w", err)
	}
	a.history = append(a.history, llm.Message{
		Role:    "assistant",
		Content: fullReply.String(),
	})
	return nil
}
