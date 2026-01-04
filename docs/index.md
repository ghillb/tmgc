---
layout: default
title: tmgc
description: "Script-first Telegram MTProto Go CLI."
hide_title: true
---

<div class="hero">
  <div>
    <p class="eyebrow">Telegram MTProto Go CLI</p>
    <h1 class="hero-title">tmgc</h1>
    <p class="lead">
      A script-first CLI on top of <strong>gotd/td</strong>. Fast QR login, stable
      output formats, and keyring-backed sessions with a file fallback.
    </p>
    <div class="hero-actions">
      <a class="btn primary" href="{{ '/getting-started.html' | relative_url }}">Get started</a>
      <a class="btn ghost" href="https://github.com/ghillb/tmgc">GitHub</a>
    </div>
    <div class="chips">
      <span class="chip">QR login</span>
      <span class="chip">Keyring storage</span>
      <span class="chip">JSON + TSV</span>
      <span class="chip">Profiles</span>
    </div>
  </div>
  <div class="terminal">
    <div class="term-header">
      <span class="term-dot"></span><span class="term-dot"></span><span class="term-dot"></span>
    </div>
    <pre><code>$ tmgc auth config set --api-id 123456 --api-hash abc123...
$ tmgc auth login
$ tmgc chat list --limit 10
$ tmgc message send @username "hello"</code></pre>
  </div>
</div>

## Highlights

- QR login with PNG fallback for mobile cameras.
- `--json` and `--plain` outputs built for scripts.
- Profiles for multi-account workflows.
- Keyring-first session storage with explicit file fallback.

## Install

```bash
go install github.com/ghillb/tmgc/cmd/tmgc@latest
```

## Quick start

```bash
tmgc auth config set --api-id 123456 --api-hash abc123...
tmgc auth login
tmgc chat list --limit 20
tmgc message send @username "hello"
```

Need full details? Start at <a href="{{ '/getting-started.html' | relative_url }}">Get Started</a>.
