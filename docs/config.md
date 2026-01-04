---
title: Config
---

## Profiles

Configs and sessions live under:

```
~/.config/tmgc/profiles/<profile>/
```

Use a different profile:

```bash
tmgc --profile work auth login
```

## Config file

Stored at `config.json` in the profile directory. Set it with:

```bash
tmgc auth config set --api-id 123456 --api-hash abc123...
```

## Session storage

Default: OS keyring. If keyring is unavailable or `TMGC_SESSION_STORE=file`,
`tmgc` uses a plaintext `session.json` file in the profile directory.

Force a backend:

```bash
export TMGC_SESSION_STORE=keyring
export TMGC_SESSION_STORE=file
```

Or set it in config:

```bash
tmgc auth config set --session-store file
```

## Environment overrides

```bash
export TMGC_API_ID=123456
export TMGC_API_HASH=abc123...
```
