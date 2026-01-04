# tmgc - Telegram MTProto Go CLI

Script-first Telegram CLI on top of `gotd/td`.

Features:

- Auth: `auth login` (QR + PNG fallback), `auth status`, `auth logout`
- Credentials: `auth config set/show`
- Chat: `chat list`, `chat history`
- Messaging: `message send` (text or `--file`)
- Search: `search messages` (global or per chat)
- Output: human, `--plain` (TSV), `--json`
- Session storage: OS keychain by default, fallback to file

Install:

```bash
go install github.com/ghillb/tmgc/cmd/tmgc@latest
```

Quick start:

```bash
tmgc auth config set --api-id 123456 --api-hash abc123...
tmgc auth login
tmgc auth login --qr-file /tmp/tmgc.png
tmgc chat list --limit 20
tmgc message send @username "hello"
tmgc message send @username --file ./photo.jpg --caption "hi"
```

Env vars are only needed if you skip `auth config set`:

```bash
export TMGC_API_ID=123456
export TMGC_API_HASH=abc123...
```

Optional: force session storage backend:

```bash
export TMGC_SESSION_STORE=keyring   # default
export TMGC_SESSION_STORE=file
```

You can also set it via config:

```bash
tmgc auth config set --session-store file
```

Docs: `docs/spec.md`

Notes: Third-party CLI, not affiliated with Telegram. QR login is recommended.
