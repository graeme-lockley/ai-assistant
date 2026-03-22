# Requirements: Configuration

All configuration is environment variables (optional `.env` via `dotenv` on CLIs).

## Server / core

| Variable | Purpose |
|----------|---------|
| `AI_ASSISTANT_BIND` | Listen address (default `:8080`) |
| `AI_ASSISTANT_WORKSPACE` | Workspace root (default `~/.ai-assistant.workspace`) |
| `AI_ASSISTANT_ROOT_DIR` | Legacy alias for workspace |
| `AI_ASSISTANT_DEFAULT_MODEL` | Default model id |
| `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` | Default `Accept` if client omits |
| `AI_ASSISTANT_BOOTSTRAP_RING2` | `false`/`0` to drop USER/MEMORY/TASKS from system prompt |
| `AI_ASSISTANT_RING2_MAX_TOKENS` | Per Ring-2 file cap |
| `AI_ASSISTANT_SYSTEM_PROMPT_MAX_TOKENS` | Global system prompt cap |
| `TAVILY_API_KEY` | `web_search` tool |

## LLM providers (pi-ai)

At least one of: `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`. Optional: `DEEPSEEK_BASE_URL`.

## Clients

| Variable | Purpose |
|----------|---------|
| `AI_ASSISTANT_SERVER_URL` | Full base URL for REPL/ask |
| `AI_ASSISTANT_SERVER_ADDR` | `host:port` fallback |
| `AI_ASSISTANT_ASK_MODEL` | Default model for `ask` |
