---
name: whatsapp
description: Send WhatsApp to third parties or search/sync history via wacli when the user explicitly asks (not for normal chat with the user).
---

# WhatsApp (wacli)

Use only when the user asks to message someone else or to search/sync WhatsApp history.

## Commands

```bash
wacli chats list --limit 20 --query "name"
wacli messages search "query" --limit 20 --chat <jid>
wacli send text --to "+15551234567" --message "Hello"
```

## Safety

- Confirm recipient and message before sending.
