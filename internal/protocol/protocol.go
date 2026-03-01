package protocol

// Request is sent from client to server (one user message per request).
// Used as JSON body for Content-Type: application/json.
type Request struct {
	Message string `json:"message"`
}

// Response is the legacy shape for a non-streamed reply (message or error).
// The HTTP API streams responses; this struct is kept for reference or future use.
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
