# Contributing

Thanks for your interest in improving `shortcut-cli`.

## Development setup

- Go `1.25+`
- GNU `make`

Common commands:

```bash
make build
go test ./...
make dist VERSION=v1.0.0 COMMIT=$(git rev-parse HEAD)
```

## Project layout

- `cmd/shortcut/main.go` — CLI entrypoint
- `internal/cli/` — curated command UX and runtime
- `internal/gen/shortcutv3/` — generated client code
- `openapi/shortcut.openapi.json` — vendored API source of truth

## Contribution guidelines

- Keep curated commands domain-oriented and easy to discover from `--help`.
- Keep raw transport details under `shortcut api ...`.
- Prefer focused changes over broad refactors.
- Do not hand-edit generated files under `internal/gen/shortcutv3/`.
- Keep docs and examples honest: if behavior changes, update the docs in the same change.

## Before opening a PR

- run `go test ./...`
- run `make build` if you changed CLI or packaging behavior
- confirm help text and examples still match the actual command output

If you are proposing a new user-facing workflow, include a short example in the docs or help output.
