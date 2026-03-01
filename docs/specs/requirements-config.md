# Requirements: Configuration

## Source

Configuration is read from **environment variables** only. No config file in the initial scope.

## Server

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_ASSISTANT_BIND` | HTTP listen address | `:8080` |
| `DEEPSEEK_API_KEY` | Deepseek API key | (required; no default) |
| `DEEPSEEK_BASE_URL` | Deepseek API base URL | `https://api.deepseek.com` |
| `DEEPSEEK_MODEL` | Model name | `deepseek-chat` |
| `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` | Default response stream type when client omits `Accept` | (optional) e.g. `text/event-stream` or `application/json` |

The server needs bind address and API key. If `DEEPSEEK_API_KEY` is empty, the server exits with an error.

## REPL

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_ASSISTANT_SERVER_ADDR` | Server address (host:port) | `127.0.0.1:8080` |
| `AI_ASSISTANT_SERVER_URL` | Full server URL (overrides SERVER_ADDR if set) | (optional) e.g. `http://127.0.0.1:8080` |
| `AI_ASSISTANT_DEFAULT_REQUEST_TYPE` | Default request body type | (optional) e.g. `application/json` or `text/plain` |
| `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` | Default Accept (stream format) | (optional) e.g. `text/event-stream` or `application/json` |

The REPL uses the server address or URL to POST each turn. It does not use any Deepseek or API keys; the server performs all LLM calls.

## Usage

- **Server**: Set `DEEPSEEK_API_KEY` (and optionally others), then run `ai-assistant server`.
- **REPL**: Optionally set `AI_ASSISTANT_SERVER_ADDR` or `AI_ASSISTANT_SERVER_URL` if the server is not on `http://127.0.0.1:8080`, then run `ai-assistant repl`.
