package protocol

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSSEWriter_WriteEvent(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSSEWriter(&buf)
	if err := sw.WriteEvent(EventToken, map[string]string{"delta": "hi"}); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "event: token") {
		t.Errorf("output missing event: token: %q", out)
	}
	if !strings.Contains(out, "data: ") {
		t.Errorf("output missing data: %q", out)
	}
	var v struct {
		Delta string `json:"delta"`
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			if err := json.Unmarshal([]byte(line[6:]), &v); err != nil {
				t.Fatalf("data line not valid JSON: %q", line)
			}
			if v.Delta != "hi" {
				t.Errorf("delta: got %q, want hi", v.Delta)
			}
			break
		}
	}
}

func TestSSEWriter_WriteEventDone(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSSEWriter(&buf)
	if err := sw.WriteEvent(EventDone, nil); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "event: done") {
		t.Errorf("output missing event: done: %q", out)
	}
}

func TestNDJSONWriter_WriteLine(t *testing.T) {
	var buf bytes.Buffer
	nw := NewNDJSONWriter(&buf)
	if err := nw.WriteLine(StreamEvent{Type: EventToken, Delta: "x"}); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var ev StreamEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		t.Fatalf("line not valid JSON: %q", line)
	}
	if ev.Type != EventToken || ev.Delta != "x" {
		t.Errorf("got %+v", ev)
	}
}

func TestNDJSONWriter_WriteLineDone(t *testing.T) {
	var buf bytes.Buffer
	nw := NewNDJSONWriter(&buf)
	if err := nw.WriteLine(StreamEvent{Type: EventDone}); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "{\"type\":\"done\"}\n") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}
