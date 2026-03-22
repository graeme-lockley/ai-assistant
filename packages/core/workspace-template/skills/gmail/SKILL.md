---
name: gmail
description: Read, search, summarize, draft, reply to, or send email via Gmail using the gog CLI when the user asks about email.
---

# Gmail (gog)

Use when the user asks to read, search, draft, send, or summarize Gmail.

## Commands

```bash
gog gmail search "query" [--max N] [--plain]
gog gmail messages <id>
gog gmail send --to "addr" --subject "X" --body-file /path
```

## Safety

- Never send without explicit user instruction; confirm recipient and body for sends.
