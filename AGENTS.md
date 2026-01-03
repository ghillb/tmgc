# AGENTS

## Project basics
- Language: Go, CLI tool.
- Build: `go build -o ./dist/tmgc ./cmd/tmgc`
- Test: `go test ./...`
- Format: `gofmt -w $(rg --files -g '*.go')`

## Storage and secrets
- Config/session/peer cache live under `~/.config/tmgc/profiles/<profile>/`.
- Never commit credentials, session files, or QR PNGs.
- Do not log passwords or tokens.

## Release rules (go-semantic-release)
- feat: → minor
- fix: → patch
- feat!: or BREAKING CHANGE: → major
- chore: etc → no release
