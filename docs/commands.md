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
| `message send <peer> <text>` | Send a message to a chat or username. |

## Search

| Command | Notes |
| --- | --- |
| `search messages <query>` | Global or per-chat search. |

## Output

Add `--json` or `--plain` to any command.
