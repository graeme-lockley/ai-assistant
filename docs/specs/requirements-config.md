# Requirements: Configuration

## Source

Configuration is read from **environment variables** only. No config file in the initial scope.

## Server

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_ASSISTANT_BIND` | HTTP listen address | `:8080` |
| `DEEPSEEK_API_KEY` | Deepseek API key | (one of Deepseek or Anthropic required) |
| `DEEPSEEK_BASE_URL` | Deepseek API base URL | `https://api.deepseek.com` |
| `DEEPSEEK_MODEL` | Default Deepseek model | `deepseek-chat` |
| `ANTHROPIC_API_KEY` | Anthropic API key (Claude) | (optional; multi-provider if both set) |
| `AI_ASSISTANT_WORKSPACE` | Workspace root path | overrides default |
| `AI_ASSISTANT_ROOT_DIR` | Workspace root path (fallback) | used if `AI_ASSISTANT_WORKSPACE` unset |
| *(default workspace)* | When neither env set | `~/.ai-assistant.workspace` |
| `TAVILY_API_KEY` | Tavily Search API key for web_search tool | (required for web_search) |
| `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` | Default response stream type when client omits `Accept` | (optional) e.g. `text/event-stream` or `application/json` |

The server needs at least one of `DEEPSEEK_API_KEY` or `ANTHROPIC_API_KEY`. File tools and session bootstrap use the workspace root (default `~/.ai-assistant.workspace`). The web_search tool requires `TAVILY_API_KEY`.

## REPL

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_ASSISTANT_SERVER_ADDR` | Server address (host:port) | `127.0.0.1:8080` |
| `AI_ASSISTANT_SERVER_URL` | Full server URL (overrides SERVER_ADDR if set) | (optional) e.g. `http://127.0.0.1:8080` |
| `AI_ASSISTANT_DEFAULT_REQUEST_TYPE` | Default request body type | (optional) e.g. `application/json` or `text/plain` |
| `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` | Default Accept (stream format) | (optional) e.g. `text/event-stream` or `application/json` |
| `AI_ASSISTANT_REPL_HISTORY_FILE` | Path to REPL input history file | `<UserConfigDir>/ai-assistant/repl_history` |
| `AI_ASSISTANT_REPL_HISTORY_MAX` | Max number of history entries to keep | `1000` |

The REPL uses the server address or URL to POST each turn. It does not use any Deepseek or API keys; the server performs all LLM calls. Input supports readline-style history (Up/Down to navigate, Left/Right to move within the line); history is persisted to the configured file and bounded by the max size.

## Usage

- **Server**: Set `DEEPSEEK_API_KEY` and/or `ANTHROPIC_API_KEY`; set `TAVILY_API_KEY` for web search. Optionally set `AI_ASSISTANT_WORKSPACE` (or `AI_ASSISTANT_ROOT_DIR`) for workspace root. Then run `ai-assistant server`.
- **REPL**: Optionally set `AI_ASSISTANT_SERVER_ADDR` or `AI_ASSISTANT_SERVER_URL` if the server is not on `http://127.0.0.1:8080`, then run `ai-assistant repl`.
