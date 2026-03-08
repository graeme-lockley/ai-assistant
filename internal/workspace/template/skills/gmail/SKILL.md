---
name: gmail
description: Read, search, summarize, draft, reply to, or send email via Gmail using the `gog` tool.
tags: gmail, email, communication
---

# Gmail

Read, search, summarize, draft, reply to, or send email via Gmail using the `gog` tool.

## when to use

- User asks to read, search, or summarize email
- User asks to draft, send, or reply to email
- User asks about attachments or inbox state

## when not to use

- User is not asking about email or sending messages
- No explicit instruction to send; never send without user approval

## Commands

```
gog gmail search "query" [--max N] [--plain]    # Search emails
gog gmail messages <id>                         # Get message content
gog gmail attachments <message-id>              # List attachments for a message
gog gmail send --to "addr" --subject "X" --body-file /path   # Send email
gog gmail send --to "addr" --subject "X" --body-file /path --attach ./file.pdf  # With attachment
```

## Search examples

```
gog gmail search "newer_than:7d is:unread" --max 10
gog gmail search "from:boss@company.com subject:meeting" --max 5
gog gmail search "is:unread newer_than:30d -category:promotions" --max 20
```

## Read email

```bash
gog gmail search "from:boss@company.com subject:meeting" --max 1 --plain
gog gmail messages <message-id>
```

## Send with attachment

```bash
echo "Please find attached the document." > /tmp/body.txt
gog gmail send --to "recipient@example.com" --subject "Document" --body-file /tmp/body.txt --attach ./file.pdf
rm /tmp/body.txt
```

Use `--body-file` to avoid shell escaping. For "send just a file", create a brief cover letter in the body file first.

## Key points

- **Attachments:** `--attach` flag; repeat for multiple files
- **Success:** Check for `message_id` in output
- **Search operators:** `is:unread`, `from:`, `to:`, `subject:`, `newer_than:Nd`, `older_than:Nd`, `has:attachment`, `label:`, `category:`

## Safety

- Require explicit recipient + message text before sending
- Confirm recipient + message before sending
- Verify attachments are the intended files
- For sensitive content (legal, financial, personal), draft first for review
- Never send without explicit user instruction; never attach wrong files; never impersonate without authorization
