# Contributing

Thanks for helping improve tmgc.

## Basics

- Use conventional commits: `type: description`
- Keep changes small and focused.
- Run tests when touching code: `go test ./...`
- One PR per fix or feature.

## Semantic versioning

tmgc uses semantic versioning driven by commit prefixes:

- `feat:` → minor
- `fix:` → patch
- `feat!:` or `BREAKING CHANGE:` → major
- `chore:`, `docs:`, `refactor:`, `test:`, `style:`, `perf:` → no release

## Docs

This project publishes GitHub Pages from `docs/`.

When adding or changing user-facing features, update the relevant docs
(`docs/` pages and `docs/spec.md`) in the same PR.

## Tests

Prefer lightweight unit tests for pure logic; no Telegram integration tests yet. Add tests when adding features or fixing bugs.
