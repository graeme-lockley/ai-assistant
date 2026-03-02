package llm

import (
	"context"
	"strings"
	"unicode/utf8"
)

// Context size limits for compression. Models like deepseek-reasoner have 131072 token context;
// we reserve headroom for the completion and system message.
const (
	DefaultMaxContextTokens = 100_000
	// MaxToolResultRunes is the maximum runes of tool result content stored in history (avoids huge web_get results).
	MaxToolResultRunes = 8_000
	// MaxReasoningRunesPerMessage is the maximum runes of ReasoningContent kept per assistant message when compressing.
	MaxReasoningRunesPerMessage = 2_000
)

// charsPerToken is a conservative estimate (OpenAI/Deepseek ~4 chars per token for English).
const charsPerToken = 4

// EstimateMessageTokens returns an approximate token count for one message (content + reasoning_content).
func EstimateMessageTokens(m *Message) int {
	n := utf8.RuneCountInString(m.Content) + utf8.RuneCountInString(m.ReasoningContent)
	for _, tc := range m.ToolCalls {
		n += utf8.RuneCountInString(tc.Arguments) + 20 // name, id overhead
	}
	return (n + charsPerToken - 1) / charsPerToken
}

// EstimateTokens returns the total estimated token count for a slice of messages (plus system message overhead).
func EstimateTokens(messages []Message) int {
	const systemOverhead = 100
	n := systemOverhead
	for i := range messages {
		n += EstimateMessageTokens(&messages[i])
	}
	return n
}

// truncateRunes truncates s to at most maxRunes runes, appending suffix if truncated.
func truncateRunes(s string, maxRunes int, suffix string) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + suffix
}

// CompressMessages returns a copy of messages that fits within maxTokens by:
// - Dropping oldest complete "turns" (user message plus all following assistant/tool messages until the next user).
// - Truncating ReasoningContent in remaining assistant messages to MaxReasoningRunesPerMessage (keeps tail).
// Always keeps at least the last user message and everything after it.
func CompressMessages(messages []Message, maxTokens int) []Message {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxContextTokens
	}
	if EstimateTokens(messages) <= maxTokens {
		// Still truncate long reasoning in place so future appends don't blow up
		out := make([]Message, len(messages))
		for i := range messages {
			out[i] = messages[i]
			if out[i].Role == "assistant" && utf8.RuneCountInString(out[i].ReasoningContent) > MaxReasoningRunesPerMessage {
				r := []rune(out[i].ReasoningContent)
				out[i].ReasoningContent = "… [earlier reasoning truncated] " + string(r[len(r)-MaxReasoningRunesPerMessage:])
			}
		}
		return out
	}

	// Find turn boundaries: each turn starts with a "user" message.
	var turnStart []int
	for i := range messages {
		if messages[i].Role == "user" {
			turnStart = append(turnStart, i)
		}
	}
	if len(turnStart) == 0 {
		return messages
	}

	// Keep dropping the oldest turn until we're under the limit.
	start := 0
	for len(turnStart) > 1 && start < turnStart[len(turnStart)-1] {
		// drop turn starting at turnStart[0] (ends just before turnStart[1])
		nextStart := turnStart[1]
		trimmed := messages[nextStart:]
		if EstimateTokens(trimmed) <= maxTokens {
			start = nextStart
			turnStart = turnStart[1:]
			break
		}
		turnStart = turnStart[1:]
		start = nextStart
	}
	// start is the first index to keep; if we only had one turn, start is still 0
	if start > 0 {
		messages = messages[start:]
	}

	// Truncate reasoning in assistant messages
	out := make([]Message, len(messages))
	for i := range messages {
		out[i] = messages[i]
		if out[i].Role == "assistant" && utf8.RuneCountInString(out[i].ReasoningContent) > MaxReasoningRunesPerMessage {
			r := []rune(out[i].ReasoningContent)
			out[i].ReasoningContent = "… [earlier reasoning truncated] " + string(r[len(r)-MaxReasoningRunesPerMessage:])
		}
	}

	// If still over (e.g. one very long turn), aggressively truncate tool and assistant content
	for EstimateTokens(out) > maxTokens {
		reduced := false
		for i := range out {
			if out[i].Role == "tool" && utf8.RuneCountInString(out[i].Content) > MaxToolResultRunes {
				out[i].Content = truncateRunes(out[i].Content, MaxToolResultRunes, "\n… [truncated for context limit]")
				reduced = true
				break
			}
			if out[i].Role == "assistant" {
				if utf8.RuneCountInString(out[i].ReasoningContent) > 500 {
					r := []rune(out[i].ReasoningContent)
					out[i].ReasoningContent = "… " + string(r[len(r)-500:])
					reduced = true
					break
				}
				if utf8.RuneCountInString(out[i].Content) > 2000 {
					out[i].Content = truncateRunes(out[i].Content, 2000, "\n… [truncated]")
					reduced = true
					break
				}
			}
		}
		if !reduced {
			break
		}
	}

	return out
}

// TruncateToolResult truncates tool result content for storage in history. Call before appending a tool message.
func TruncateToolResult(content string) string {
	s := strings.TrimSpace(content)
	if utf8.RuneCountInString(s) <= MaxToolResultRunes {
		return s
	}
	return truncateRunes(s, MaxToolResultRunes, "\n… [result truncated for context limit]")
}

// Summarizer returns a short summary of the given messages (e.g. for context compression).
// If summarization fails, the caller may fall back to dropping those messages.
type Summarizer func(ctx context.Context, messages []Message) (summary string, err error)

const summarizerSystemPrompt = "You are a conversation summarizer. Summarize the following conversation in 2-5 concise sentences. Preserve: key facts, decisions, user preferences, and any context the assistant will need to continue. Omit routine greetings and tool implementation details; focus on what was learned or decided."

// formatMessagesForSummary flattens messages into a single string for the summarizer.
func formatMessagesForSummary(messages []Message) string {
	var b strings.Builder
	for _, m := range messages {
		switch m.Role {
		case "user":
			b.WriteString("User: ")
			b.WriteString(m.Content)
			b.WriteString("\n")
		case "assistant":
			if m.Content != "" {
				b.WriteString("Assistant: ")
				b.WriteString(m.Content)
				b.WriteString("\n")
			}
			if m.ReasoningContent != "" {
				// Keep reasoning very short in summary input
				r := []rune(m.ReasoningContent)
				if len(r) > 300 {
					r = r[len(r)-300:]
				}
				b.WriteString("(reasoning: ")
				b.WriteString(string(r))
				b.WriteString(")\n")
			}
		case "tool":
			c := m.Content
			if utf8.RuneCountInString(c) > 500 {
				c = truncateRunes(c, 500, "…")
			}
			b.WriteString("Tool result: ")
			b.WriteString(c)
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}

// SummarizerFromCompleter returns a Summarizer that uses the Completer to summarize the given messages.
// Useful when the same LLM client should be used for both chat and summarization.
func SummarizerFromCompleter(c Completer) Summarizer {
	return func(ctx context.Context, messages []Message) (string, error) {
		if len(messages) == 0 {
			return "", nil
		}
		formatted := formatMessagesForSummary(messages)
		summaryMsgs := []Message{
			{Role: "system", Content: summarizerSystemPrompt},
			{Role: "user", Content: formatted},
		}
		return c.Complete(ctx, summaryMsgs)
	}
}

// CompressMessagesWithSummarizer compresses messages to fit within maxTokens. When turns would be dropped,
// summarizer is called on the dropped messages and the result is prepended as a system message so context
// is preserved instead of lost. If summarizer is nil or returns an error, falls back to dropping without summary.
func CompressMessagesWithSummarizer(ctx context.Context, messages []Message, maxTokens int, summarizer Summarizer) ([]Message, error) {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxContextTokens
	}
	if EstimateTokens(messages) <= maxTokens {
		// Same as CompressMessages: truncate long reasoning only
		out := make([]Message, len(messages))
		for i := range messages {
			out[i] = messages[i]
			if out[i].Role == "assistant" && utf8.RuneCountInString(out[i].ReasoningContent) > MaxReasoningRunesPerMessage {
				r := []rune(out[i].ReasoningContent)
				out[i].ReasoningContent = "… [earlier reasoning truncated] " + string(r[len(r)-MaxReasoningRunesPerMessage:])
			}
		}
		return out, nil
	}

	var turnStart []int
	for i := range messages {
		if messages[i].Role == "user" {
			turnStart = append(turnStart, i)
		}
	}
	if len(turnStart) == 0 {
		return messages, nil
	}

	// Find how many turns we must drop to get under the limit
	start := 0
	const summaryBudget = 500
	maxWithoutSummary := maxTokens - summaryBudget
	for len(turnStart) > 1 && start < turnStart[len(turnStart)-1] {
		nextStart := turnStart[1]
		trimmed := messages[nextStart:]
		if EstimateTokens(trimmed) <= maxWithoutSummary {
			start = nextStart
			turnStart = turnStart[1:]
			break
		}
		turnStart = turnStart[1:]
		start = nextStart
	}

	if start == 0 {
		// No turns dropped; still apply truncation of long content
		out := make([]Message, len(messages))
		for i := range messages {
			out[i] = messages[i]
			if out[i].Role == "assistant" && utf8.RuneCountInString(out[i].ReasoningContent) > MaxReasoningRunesPerMessage {
				r := []rune(out[i].ReasoningContent)
				out[i].ReasoningContent = "… [earlier reasoning truncated] " + string(r[len(r)-MaxReasoningRunesPerMessage:])
			}
		}
		for EstimateTokens(out) > maxTokens {
			reduced := false
			for i := range out {
				if out[i].Role == "tool" && utf8.RuneCountInString(out[i].Content) > MaxToolResultRunes {
					out[i].Content = truncateRunes(out[i].Content, MaxToolResultRunes, "\n… [truncated for context limit]")
					reduced = true
					break
				}
				if out[i].Role == "assistant" {
					if utf8.RuneCountInString(out[i].ReasoningContent) > 500 {
						r := []rune(out[i].ReasoningContent)
						out[i].ReasoningContent = "… " + string(r[len(r)-500:])
						reduced = true
						break
					}
					if utf8.RuneCountInString(out[i].Content) > 2000 {
						out[i].Content = truncateRunes(out[i].Content, 2000, "\n… [truncated]")
						reduced = true
						break
					}
				}
			}
			if !reduced {
				break
			}
		}
		return out, nil
	}

	dropped := messages[:start]
	kept := messages[start:]

	// Reserve tokens for a summary message (generous)
	if summarizer != nil && EstimateTokens(kept)+summaryBudget <= maxTokens {
		summary, err := summarizer(ctx, dropped)
		if err == nil && summary != "" {
			summaryMsg := Message{
				Role:    "system",
				Content: "Previous conversation summary:\n" + summary,
			}
			kept = append([]Message{summaryMsg}, kept...)
			if EstimateTokens(kept) > maxTokens {
				// Summary pushed us over; drop it and just use kept
				kept = messages[start:]
			}
		}
	}

	out := make([]Message, len(kept))
	for i := range kept {
		out[i] = kept[i]
		if out[i].Role == "assistant" && utf8.RuneCountInString(out[i].ReasoningContent) > MaxReasoningRunesPerMessage {
			r := []rune(out[i].ReasoningContent)
			out[i].ReasoningContent = "… [earlier reasoning truncated] " + string(r[len(r)-MaxReasoningRunesPerMessage:])
		}
	}

	for EstimateTokens(out) > maxTokens {
		reduced := false
		for i := range out {
			if out[i].Role == "tool" && utf8.RuneCountInString(out[i].Content) > MaxToolResultRunes {
				out[i].Content = truncateRunes(out[i].Content, MaxToolResultRunes, "\n… [truncated for context limit]")
				reduced = true
				break
			}
			if out[i].Role == "assistant" {
				if utf8.RuneCountInString(out[i].ReasoningContent) > 500 {
					r := []rune(out[i].ReasoningContent)
					out[i].ReasoningContent = "… " + string(r[len(r)-500:])
					reduced = true
					break
				}
				if utf8.RuneCountInString(out[i].Content) > 2000 {
					out[i].Content = truncateRunes(out[i].Content, 2000, "\n… [truncated]")
					reduced = true
					break
				}
			}
		}
		if !reduced {
			break
		}
	}

	return out, nil
}
