# tmgc CLI spec (v0)

This spec defines the CLI surface, output contract, and minimal behavior. It favors
stable machine output (`--json`/`--plain`) and clean human output by default.

## Global

Usage:

```
tmgc [global flags] <command> [args]
```

Global flags:

- `--config <path>`: config file path (default: `~/.config/tmgc/profiles/<profile>/config.json`)
- `--profile <name>`: profile name (default: `default`)
- `--timeout <dur>`: request timeout (default: `15s`)
- `--json`: JSON output
- `--plain`: line-oriented output (TSV)
- `--no-color`: disable colors

Environment overrides:

- `TMGC_API_ID`
- `TMGC_API_HASH`

Config storage:

- Config: `~/.config/tmgc/profiles/<profile>/config.json`
- Session: `~/.config/tmgc/profiles/<profile>/session.json`
- Peer cache: `~/.config/tmgc/profiles/<profile>/peers.json`

## Output

- **Human** (default): tabular output on stdout.
- **Plain** (`--plain`): stable TSV on stdout (tabs preserved), ideal for piping.
- **JSON** (`--json`): structured JSON on stdout.
- Progress/warnings go to stderr.

## Peer references

Accepted peer inputs:

- `u<id>`: user (e.g., `u123456`)
- `c<id>`: chat (group) (e.g., `c123456`)
- `ch<id>`: channel / supergroup (e.g., `ch123456`)
- `@username`, `t.me/username`, `tg://resolve?domain=...`
- phone number (E.164 or with separators)

## Auth

`tmgc auth login` defaults to QR login (fastest, safest). Code login is supported
as a fallback for accounts that cannot scan a QR.

Notes:

- Telegram may choose non-SMS code types (app, email, Fragment, etc.).
- If 2FA is enabled, a password prompt appears after code entry.
- QR codes expire quickly; the CLI will refresh and re-print when needed.

## Commands

### `auth`

```
tmgc auth login [--method qr|code] [--phone ...] [--api-id ...] [--api-hash ...] [--bot-token ...]
tmgc auth login [--qr-file /path/to/qr.png]
tmgc auth status
tmgc auth logout
tmgc auth config set --api-id ... --api-hash ...
tmgc auth config show
```

`auth login` outputs an `AuthStatus` JSON object or a summary table.

### `chat`

```
tmgc chat list [--limit 50]
tmgc chat history <peer> [--limit 20] [--since RFC3339]
```

#### `chat list`

Output (JSON):

```json
[
  {
    "peer_id": 123456,
    "peer_ref": "u123456",
    "peer_type": "user",
    "title": "Jane Doe",
    "username": "jane",
    "unread_count": 2,
    "last_message_id": 9876,
    "pinned": false
  }
]
```

#### `chat history`

Output (JSON):

```json
[
  {
    "id": 42,
    "date": "2026-01-03T20:15:00Z",
    "text": "hello",
    "from_peer_id": 123456,
    "peer_id": 123456,
    "out": true,
    "service": false
  }
]
```

### `message`

```
tmgc message send <peer> <text> [--reply <id>] [--silent]
```

Output (JSON):

```json
{
  "ok": true,
  "message_id": 123,
  "updates_type": "*tg.Updates"
}
```

### `search`

```
tmgc search messages <query> [--chat <peer>] [--limit 20]
```

Output shape matches `chat history`.

## Scope

v0 is scoped to:

- Auth (QR + code), status, logout
- Chat list + history
- Send text messages
- Search messages (global or per chat)

Non-goals for v0:

- Media upload/download
- Full offline sync
- Multi-account token management beyond profiles
