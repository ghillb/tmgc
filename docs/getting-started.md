---
title: Get Started
---

## 1) Install

```bash
go install github.com/ghillb/tmgc/cmd/tmgc@latest
```

## 2) Set API credentials

Preferred (writes config):

```bash
tmgc auth config set --api-id 123456 --api-hash abc123...
```

Alternate (env vars):

```bash
export TMGC_API_ID=123456
export TMGC_API_HASH=abc123...
```

## 3) Login

```bash
tmgc auth login
```

If the QR camera cannot focus, render a PNG:

```bash
tmgc auth login --qr-file /tmp/tmgc.png
```

## 4) First commands

```bash
tmgc chat list --limit 20
tmgc chat history <chat_id> --limit 30
tmgc message send @username "hello"
tmgc message send @username --file ./photo.jpg --caption "hi"
tmgc message send @username "later" --schedule 2026-01-05T09:30:00Z
```

## Output modes

- Human (default)
- `--plain` for TSV
- `--json` for machine consumption

Example:

```bash
tmgc chat list --json | jq '.chats[0]'
```
