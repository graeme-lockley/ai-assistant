---
description: Read, search, summarize, draft, reply to, or send email via Gmail using the `gog` tool.
name: gmail-agent
---

# Gmail Agent Skill

## Commands
```
gog gmail search "query" [--max N] [--plain]    # Search emails
gog gmail messages <id>                        # Get message content
gog gmail attachments <message-id>              # List attachments for a message
gog gmail send --to "addr" --subject "X" --body-file /path   # Send email
gog gmail send --to "addr" --subject "X" --body-file /path --attach ./file.pdf  # With attachment
```

## Search Examples
```
gog gmail search "newer_than:7d is:unread" --max 10
gog gmail search "from:boss@company.com subject:meeting" --max 5
gog gmail search "is:unread newer_than:30d -category:promotions" --max 20
```

## Read Email
```bash
# First search to get message ID
gog gmail search "from:boss@company.com subject:meeting" --max 1 --plain

# Then read the message content
gog gmail messages <message-id>
```

## Read Attachments
```bash
# List attachments for a message
gog gmail attachments <message-id>
```

## Send with Attachment
```bash
# Create body file
echo "Please find attached the document." > /tmp/body.txt

# Send
gog gmail send --to "recipient@example.com" --subject "Document" --body-file /tmp/body.txt --attach ./file.pdf

# Cleanup
rm /tmp/body.txt
```

If asked to "send just a file" (no body content), create a brief cover letter:
```bash
cat > /tmp/body.txt << 'EOF'
Hi,

Please find attached the requested document.

Best regards
EOF

gog gmail send --to "recipient@example.com" --subject "Document" --body-file /tmp/body.txt --attach ./file.pdf
```

## Sending to Multiple Recipients
```bash
gog gmail send --to "person1@example.com" --subject "Subject" --body-file /tmp/body.txt
gog gmail send --to "person2@example.com" --subject "Subject" --body-file /tmp/body.txt
```

## Key Points
- **Attachments**: Use `--attach` flag; repeat for multiple files
- **Body**: Use `--body-file` to avoid shell escaping issues
- **Success**: Check for `message_id` in output
- **Date queries**: `newer_than:30d`, `older_than:7d`

## Search Operators
- `is:unread`, `is:read`
- `from:`, `to:`, `subject:`
- `newer_than:Nd`, `older_than:Nd`
- `has:attachment`, `label:`, `category:`

## Never
- Send without explicit user instruction
- Attach wrong files
- Impersonate without authorization

## Safety
- Require explicit recipient + message text before sending
- Confirm recipient + message before sending
- If anything is ambiguous, ask a clarifying question
- Verify attachments are the intended files
- For sensitive content (legal, financial, personal), draft first for review
