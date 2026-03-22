---
name: cron-scheduler
description: Configure cron jobs for overnight memory consolidation and scheduled assistant prompts using ai-assistant CLI packages.
---

# Cron scheduler

## Memory consolidation (nightly)

Run the consolidate tool after the day ends so `memory/daily/` and session logs roll into `MEMORY.md`:

```bash
0 2 * * * cd /path/to/ai-assistant && AI_ASSISTANT_WORKSPACE=$HOME/.ai-assistant.workspace npx tsx packages/consolidate/src/index.ts
```

Or after global install / `npm link`, use the `ai-consolidate` binary if exposed in PATH.

Set API keys in the crontab environment or in a small wrapper script that sources `.env`.

## Scheduled prompts (optional)

Use `ask` to run a one-shot instruction on a schedule:

```bash
0 8 * * * AI_ASSISTANT_SERVER_URL=http://127.0.0.1:8080 npx tsx packages/ask/src/index.ts "Summarize today's priorities"
```

Ensure the **server** is running if you use `ask` against HTTP.

## Notes

- Prefer absolute paths in crontab.
- Redirect stdout/stderr to a log file for debugging: `>>$HOME/.ai-assistant.cron.log 2>&1`
