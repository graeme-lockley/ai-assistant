---
name: whatsapp
description: Send WhatsApp messages to other people or search/sync WhatsApp history via the wacli CLI (not for normal user chats).
tags: whatsapp, wacli, messaging, communication
homepage: https://wacli.sh
---

# WhatsApp (wacli)

Send WhatsApp messages to other people or search/sync WhatsApp history via the `wacli` CLI. Use only when the user explicitly asks you to message someone else on WhatsApp or to sync/search WhatsApp history. Do not use for normal user chats; if the user is chatting with you on WhatsApp, do not reach for this skill unless they ask you to contact a third party.

## when to use

- User explicitly asks to message someone else on WhatsApp
- User asks to sync or search WhatsApp history

## when not to use

- Routine user chats (conversation is already over WhatsApp)
- User has not asked to contact a third party or access history

## Safety

- Require explicit recipient + message text
- Confirm recipient + message before sending
- If anything is ambiguous, ask a clarifying question

## Auth and sync

- `wacli auth` (QR login + initial sync)
- `wacli sync --follow` (continuous sync)
- `wacli doctor`

## Find chats and messages

- `wacli chats list --limit 20 --query "name or number"`
- `wacli messages search "query" --limit 20 --chat <jid>`
- `wacli messages search "invoice" --after 2025-01-01 --before 2025-12-31`

## History backfill

- `wacli history backfill --chat <jid> --requests 2 --count 50`

## Send

- Text: `wacli send text --to "+14155551212" --message "Hello! Are you free at 3pm?"`
- Group: `wacli send text --to "1234567890-123456789@g.us" --message "Running 5 min late."`
- File: `wacli send file --to "+14155551212" --file /path/agenda.pdf --caption "Agenda"`

## Notes

- Store dir: `~/.wacli` (override with `--store`)
- Use `--json` for machine-readable output when parsing
- Backfill requires phone online; results are best-effort
- JIDs: direct chats `<number>@s.whatsapp.net`; groups `<id>@g.us` (use `wacli chats list` to find)
