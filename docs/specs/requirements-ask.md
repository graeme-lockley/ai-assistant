# Requirements: Ask

## CLI

`ai-assistant-ask <instruction...> [--model <id>] [--agent <name>]`

## Behavior

- `POST /` with `Accept: application/json` (NDJSON).
- Prints a single JSON object to stdout: `{ entries, model, session_id, tokens }` (same shape as the legacy Go command).
- Non-zero exit on error; stderr unused for machine-readable errors (JSON on stdout).

## Environment

Uses `AI_ASSISTANT_SERVER_URL` / `AI_ASSISTANT_SERVER_ADDR` and optional `AI_ASSISTANT_ASK_MODEL`.
