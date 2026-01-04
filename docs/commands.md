---
title: Commands
---

## Auth

| Command | Notes |
| --- | --- |
| `auth login` | QR login with optional `--qr-file` PNG output. |
| `auth status` | Show session status. |
| `auth logout` | Clear session for the profile. |
| `auth config set` | Persist API ID/hash and defaults. |
| `auth config show` | Display current config. |

## Chats

| Command | Notes |
| --- | --- |
| `chat list` | List recent dialogs. |
| `chat history <chat_id>` | Read history from a chat. |

## Messaging

| Command | Notes |
| --- | --- |
| `message send <peer> <text>` | Send a text message. |
| `message send <peer> --file <path> [--caption "text"]` | Upload media or document (auto-detected). |
| `message send <peer> ... --schedule <when>` | Schedule a message (RFC3339 or unix seconds). |

## Search

| Command | Notes |
| --- | --- |
| `search messages <query>` | Global or per-chat search. |

## Output

Add `--json` or `--plain` to any command.
