# Agent Guidelines for ai-assistant

This is a Go project (go 1.22) implementing an HTTP server and REPL client for an AI assistant backed by Deepseek.

## Build, Lint, and Test Commands

### Build
```bash
go build -o ai-assistant ./cmd/ai-assistant
```

### Run the server
```bash
export DEEPSEEK_API_KEY=your-key
./ai-assistant server
```

Optional environment variables:
- `BIND_ADDR=:8080` - server bind address (default `:8080`)
- `DEEPSEEK_BASE_URL` - override API URL
- `DEEPSEEK_MODEL` - override model (default `deepseek-chat`)
- `AI_ASSISTANT_ROOT_DIR` - root directory for file tools and exec_bash

### Run the REPL client
```bash
./ai-assistant repl
```

Optional: `AI_ASSISTANT_SERVER_ADDR` or `AI_ASSISTANT_SERVER_URL`

### Run all tests
```bash
go test ./...
```

### Run a single test
```bash
go test -v -run=TestAgent_RespondStream_SendsChunksAndAppendsToHistory ./internal/agent/
```

### Run tests with verbose output
```bash
go test -v ./...
```

### Run benchmarks (if any)
```bash
go test -bench=. ./...
```

### Check for issues (go vet)
```bash
go vet ./...
```

### Format code
```bash
go fmt ./...
```

## Code Style Guidelines

### General Principles
- Follow standard Go idioms and conventions
- Keep code simple and readable
- Use meaningful names - clarity over cleverness

### Naming Conventions
- **Packages**: lowercase, short, no underscores (e.g., `agent`, `llm`, `tools`)
- **Types/Interfaces**: PascalCase (e.g., `Agent`, `Runner`, `StreamCompleter`)
- **Functions/Methods**: PascalCase (e.g., `New`, `RespondStream`, `Run`)
- **Variables/Constants**: camelCase or mixedCaps
- **Constants**: PascalCase or camelCase depending on visibility
- **Unexported (private)**: camelCase (e.g., `llmClient`, `runner`)
- **Acronyms**: maintain original casing (e.g., `HTTP`, `URL` not `Http`, `Url`)
- **Interface names**: descriptive noun with `er` suffix where appropriate (e.g., `Reader`, `Runner`)

### Imports
- Group imports: standard library first, then third-party
- Use blank import (`_`) only when necessary for side effects
- Avoid importing packages you don't use

```go
import (
    "context"
    "fmt"
    "log"
    "strings"

    "github.com/graemelockley/ai-assistant/internal/llm"
    "github.com/graemelockley/ai-assistant/internal/tools"
)
```

### Formatting
- Use `go fmt` for automatic formatting
- No trailing commas
- One var declaration per line (avoid grouping unless initializing together)
- Group related const declarations

### Types and Declarations
- Use explicit type annotations when clarity improves
- Prefer concrete types over interfaces unless polymorphism is needed
- Return interfaces from constructors when appropriate (e.g., `NewRunner(...) (Runner, error)`)
- Use pointers (`*`) for mutable receivers and large structs

### Error Handling
- Return errors with context using `fmt.Errorf("context: %w", err)` pattern
- Check errors immediately after calling functions
- Don't ignore errors with `_` unless explicitly documented
- Use sentinel errors for known error conditions when appropriate

```go
if err != nil {
    return fmt.Errorf("agent respond: %w", err)
}
```

### Logging
- Use `log.Printf` for operational logging (not structured JSON logging)
- Prefix log messages with meaningful tags in brackets: `[context]`, `[tool]`, `[session]`
- Avoid logging sensitive data (API keys, secrets)

### Testing
- Test files: `*_test.go` in same package as implementation
- Use table-driven tests for multiple test cases
- Name test functions: `Test<Method>_<Scenario>` (e.g., `TestAgent_RespondStream_SendsChunksAndAppendsToHistory`)
- Use `t.Fatalf` or `t.Fatal` for fatal assertions, `t.Errorf` for non-fatal
- Mock external dependencies with local structs

```go
type mockStreamCompleter struct {
    reply string
    err   error
}

func (m *mockStreamCompleter) CompleteStream(ctx context.Context, ...) error {
    // implementation
}
```

### HTTP Handlers
- Use `http.HandlerFunc` with helper constructors
- Return appropriate HTTP status codes
- Set Content-Type headers explicitly
- Use context for cancellation/timeouts

### Struct Tags
- Use JSON tags for serialization: `json:"field_name"`
- Group tags on same line when possible

### Context
- Pass `context.Context` as first argument to functions that may timeout or be cancelled
- Use `context.Background()` for top-level operations in tests

### Comments
- Document exported functions and types
- Use complete sentences with proper punctuation
- Explain "why", not just "what"

### Performance Considerations
- Reuse buffers with `strings.Builder` for string concatenation in loops
- Close response bodies (`defer resp.Body.Close()`)
- Use buffered I/O where appropriate
