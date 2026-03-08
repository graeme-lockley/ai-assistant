package protocol

import (
	"strings"
	"testing"
)

func TestParseRequestBody_JSON(t *testing.T) {
	body := `{"message": "hello world"}`
	got, _, err := ParseRequestBody(strings.NewReader(body), ContentTypeJSON)
	if err != nil {
		t.Fatalf("ParseRequestBody: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want hello world", got)
	}
}

func TestParseRequestBody_JSONEmptyMessage(t *testing.T) {
	body := `{"message": ""}`
	got, _, err := ParseRequestBody(strings.NewReader(body), ContentTypeJSON)
	if err != nil {
		t.Fatalf("ParseRequestBody: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestParseRequestBody_TextPlain(t *testing.T) {
	body := "hello from text/plain"
	got, _, err := ParseRequestBody(strings.NewReader(body), ContentTypeText)
	if err != nil {
		t.Fatalf("ParseRequestBody: %v", err)
	}
	if got != body {
		t.Errorf("got %q, want %q", got, body)
	}
}

func TestParseRequestBody_UnsupportedType(t *testing.T) {
	_, _, err := ParseRequestBody(strings.NewReader("x"), "application/octet-stream")
	if err == nil {
		t.Fatal("expected error for unsupported content type")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported: %v", err)
	}
}
