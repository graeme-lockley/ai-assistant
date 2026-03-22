# Requirements: Consolidate

## Purpose

Nightly (or manual) job to move **short-term** material into **long-term** `MEMORY.md`:

- Reads `memory/daily/YYYY-MM-DD.md` for **yesterday** (relative to run time).
- Reads `logs/YYYY-MM-DD-*.md` for the same calendar day.
- Calls `completeSimple` (pi-ai) with a fixed consolidation system prompt.
- Writes `memory/weekly/YYYY-MM-DD-summary.md` and appends a dated section to `MEMORY.md`.

## CLI

`ai-assistant-consolidate [--workspace <path>]`

Default workspace from `AI_ASSISTANT_WORKSPACE` or `~/.ai-assistant.workspace`.

## Requirements

- Same API keys as the server (whichever provider `defaultModelRefFromEnv` selects).
- Idempotent enough for cron: safe to run when there is no input (no-op log line).

## Future

- Swap in pluggable `SessionStorage`-style interface for memory backends.
- Richer staging (weekly → monthly) as separate invocations or flags.
