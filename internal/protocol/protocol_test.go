package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRequest_JSONRoundTrip(t *testing.T) {
	req := &Request{Message: "hello"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Message != req.Message {
		t.Errorf("got Message %q, want %q", got.Message, req.Message)
	}
}

func TestRequest_EmptyMessage(t *testing.T) {
	req := &Request{Message: ""}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Message != "" {
		t.Errorf("got Message %q, want empty", got.Message)
	}
}

func TestResponse_JSONRoundTrip(t *testing.T) {
	resp := &Response{Message: "assistant reply", Error: ""}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Message != resp.Message {
		t.Errorf("got Message %q, want %q", got.Message, resp.Message)
	}
}

func TestResponse_ErrorJSONRoundTrip(t *testing.T) {
	resp := &Response{Error: "something went wrong"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Error != resp.Error {
		t.Errorf("got Error %q, want %q", got.Error, resp.Error)
	}
}

func TestParseRequestBody_UsesRequestStruct(t *testing.T) {
	// ParseRequestBody for application/json should unmarshal into the same shape as Request
	body := `{"message": "test"}`
	msg, err := ParseRequestBody(bytes.NewReader([]byte(body)), ContentTypeJSON)
	if err != nil {
		t.Fatalf("ParseRequestBody: %v", err)
	}
	if msg != "test" {
		t.Errorf("got %q, want test", msg)
	}
}
