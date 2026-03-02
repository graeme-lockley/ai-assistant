package llm

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEstimateTokens(t *testing.T) {
	// ~1 token per 4 chars + system overhead
	empty := []Message{}
	if got := EstimateTokens(empty); got < 50 || got > 200 {
		t.Errorf("EstimateTokens(empty): got %d, want ~100 (system overhead)", got)
	}
	short := []Message{
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello"},
	}
	got := EstimateTokens(short)
	// "Hi" + "Hello" = 7 chars -> ~2 tokens, + system ~100
	if got < 50 || got > 150 {
		t.Errorf("EstimateTokens(short): got %d", got)
	}
	// One long message: 40k chars -> ~10k tokens
	long := []Message{
		{Role: "user", Content: strings.Repeat("x", 40_000)},
	}
	gotLong := EstimateTokens(long)
	if gotLong < 9_000 || gotLong > 11_000 {
		t.Errorf("EstimateTokens(long): got %d, want ~10k", gotLong)
	}
}

func TestCompressMessages_UnderLimit_TruncatesReasoning(t *testing.T) {
	// Slightly over MaxReasoningRunesPerMessage so we truncate reasoning
	reasoning := strings.Repeat("r", MaxReasoningRunesPerMessage+500)
	msgs := []Message{
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hi there", ReasoningContent: reasoning},
	}
	out := CompressMessages(msgs, 100_000)
	if len(out) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(out))
	}
	if r := utf8.RuneCountInString(out[1].ReasoningContent); r > MaxReasoningRunesPerMessage+100 {
		t.Errorf("reasoning should be truncated: got %d runes", r)
	}
	if !strings.Contains(out[1].ReasoningContent, "truncated") {
		t.Errorf("expected truncation marker in reasoning")
	}
}

func TestCompressMessages_DropsOldestTurns(t *testing.T) {
	// Build history that exceeds limit: many user+assistant turns
	var msgs []Message
	for i := 0; i < 20; i++ {
		msgs = append(msgs, Message{Role: "user", Content: "User message number " + strings.Repeat("x", 100)})
		msgs = append(msgs, Message{Role: "assistant", Content: strings.Repeat("a", 2000)})
	}
	tokensBefore := EstimateTokens(msgs)
	if tokensBefore < 10_000 {
		t.Fatalf("test setup: expected >10k tokens, got %d", tokensBefore)
	}
	out := CompressMessages(msgs, 5_000)
	tokensAfter := EstimateTokens(out)
	if tokensAfter > 6_000 {
		t.Errorf("after compress: got %d tokens, want <= 6000", tokensAfter)
	}
	// Should keep at least last user + assistant
	userCount := 0
	for _, m := range out {
		if m.Role == "user" {
			userCount++
		}
	}
	if userCount < 1 {
		t.Error("expected at least one user message kept")
	}
}

func TestTruncateToolResult(t *testing.T) {
	short := "small result"
	if got := TruncateToolResult(short); got != short {
		t.Errorf("TruncateToolResult(short): got %q", got)
	}
	long := strings.Repeat("x", MaxToolResultRunes+1000)
	got := TruncateToolResult(long)
	if r := utf8.RuneCountInString(got); r > MaxToolResultRunes+50 {
		t.Errorf("TruncateToolResult(long): got %d runes", r)
	}
	if !strings.Contains(got, "truncated") {
		t.Errorf("expected truncation marker")
	}
}

func TestCompressMessagesWithSummarizer_SummarizesWhenDropping(t *testing.T) {
	// Build history over limit so we drop the first turn
	var msgs []Message
	for i := 0; i < 15; i++ {
		msgs = append(msgs, Message{Role: "user", Content: "User message " + strings.Repeat("x", 500)})
		msgs = append(msgs, Message{Role: "assistant", Content: strings.Repeat("a", 2000)})
	}
	ctx := context.Background()
	summarizer := func(ctx context.Context, dropped []Message) (string, error) {
		return "User asked several questions. Assistant replied with long answers.", nil
	}
	out, err := CompressMessagesWithSummarizer(ctx, msgs, 5_000, summarizer)
	if err != nil {
		t.Fatalf("CompressMessagesWithSummarizer: %v", err)
	}
	// Should start with a system summary message
	if len(out) < 1 || out[0].Role != "system" {
		t.Fatalf("expected first message to be system summary, got %d messages, first role %q", len(out), out[0].Role)
	}
	if !strings.Contains(out[0].Content, "Previous conversation") || !strings.Contains(out[0].Content, "User asked") {
		t.Errorf("expected summary in first message, got %q", out[0].Content)
	}
	if EstimateTokens(out) > 6_000 {
		t.Errorf("compressed messages over token limit: %d", EstimateTokens(out))
	}
}

func TestCompressMessagesWithSummarizer_NilSummarizer_DropsOnly(t *testing.T) {
	msgs := []Message{
		{Role: "user", Content: strings.Repeat("x", 50_000)},
	}
	out, err := CompressMessagesWithSummarizer(context.Background(), msgs, 1_000, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still have the one message (truncation only, no turns to drop)
	if len(out) != 1 {
		t.Errorf("expected 1 message when nil summarizer, got %d", len(out))
	}
}
