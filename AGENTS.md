# AGENTS

## Project basics
- Language: Go, CLI tool.
- Build: `go build -o ./dist/tmgc ./cmd/tmgc`
- Test: `go test ./...`
- Format: `gofmt -w $(rg --files -g '*.go')`
- Contributing guide: `CONTRIBUTING.md`

## Storage and secrets
- Config/session/peer cache live under `~/.config/tmgc/profiles/<profile>/`.
- Never commit credentials, session files, or QR PNGs.
- Do not log passwords or tokens.

## Release rules (go-semantic-release)
- feat: → minor
- fix: → patch
- feat!: or BREAKING CHANGE: → major
- chore: etc → no release

## Docs updates
- Update `docs/` (GitHub Pages) whenever a user-facing feature changes.
