[![CI](https://github.com/nazar256/shortcut-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/nazar256/shortcut-cli/actions/workflows/ci.yml)
[![Release](https://github.com/nazar256/shortcut-cli/actions/workflows/release.yml/badge.svg)](https://github.com/nazar256/shortcut-cli/actions/workflows/release.yml)
[![Go 1.25+](https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go)](https://go.dev/)

# shortcut-cli

Binary name: `shortcut`

`shortcut-cli` is a single-binary Go CLI for the official Shortcut REST API. It is built for AI agents, automation, and terminal-first engineers who need Shortcut access when MCP is unavailable, unnecessary, or too limited for the task.

It includes both read and write operations. Use curated read-only commands by default when you only need lookup or reporting flows.

## Why this exists

- **MCP fallback for agents**: use Shortcut from a normal shell when MCP is unavailable.
- **Help-first discovery**: the command tree is designed to be explored through `--help` and built-in docs.
- **Automation-friendly output**: concise text by default, stable JSON when a script or agent needs machine-readable output.
- **Full API escape hatch**: curated top-level commands for common workflows, plus raw `shortcut api ...` coverage from the vendored OpenAPI spec.

## Install

### From a published GitHub Release

Use this path once the repository has a published release on the [GitHub Releases page](https://github.com/nazar256/shortcut-cli/releases):

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/latest/download/install.sh | sh
```

If the installer chooses a directory that is not already on `PATH`, it prints the export command to run before you invoke `shortcut` from a new shell.

Install a specific version:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/download/v1.0.0/install.sh | sh -s -- --version v1.0.0
```

The installer:

- detects `linux` or `darwin`
- detects `amd64` or `arm64`
- downloads the matching archive from GitHub Releases
- verifies the SHA256 checksum before extraction
- prefers a writable canonical directory already in `PATH`
- avoids common language-managed bin directories such as nvm, pnpm, Cargo, and Go bin paths
- falls back to `~/.local/bin`, then `~/bin`
- supports `--install-dir` to override the default destination

Published release artifacts target:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

Download binaries directly from the [GitHub Releases page](https://github.com/nazar256/shortcut-cli/releases).

### From source

```bash
go install github.com/nazar256/shortcut-cli/cmd/shortcut@latest
```

This path works immediately from the repository state, even before the first public release is published.

Building from source requires Go `1.25+`.

Note: source installs do not inject release build metadata, so `shortcut version` will typically report `dev` / `unknown` values unless you build with ldflags.

## Authentication and config

Set your Shortcut API token:

```bash
export SHORTCUT_API_TOKEN="your_token"
```

Or copy `.env.example` to `.env` in the directory where you run the CLI.

If you are not working from a repository checkout, create `.env` manually:

```env
SHORTCUT_API_TOKEN=your_token
SHORTCUT_TIMEOUT=20s
```

Supported config:

- `SHORTCUT_API_TOKEN`
- `SHORTCUT_BASE_URL`
- `SHORTCUT_TIMEOUT`

Commands such as `shortcut --help`, `shortcut version`, and `shortcut docs summary` work without a token.

## Discover the CLI from help

```bash
shortcut --help
shortcut search --help
shortcut stories --help
shortcut docs summary
shortcut api --help
```

The intended flow is: start with top-level help, inspect a command group, then switch to JSON output or the raw `api` surface when you need exact machine-readable behavior.

## Real examples

### Read-only human-friendly flows

```bash
shortcut me
shortcut workflows list
shortcut search stories --query 'owner:example-user is:started' --detail slim
shortcut stories get 123 --with-comments
```

### Machine-readable output for scripts and agents

```bash
shortcut me -o json
shortcut docs summary -o json
shortcut workflows list -o json
shortcut search stories --query 'type:bug' --detail slim -o json
```

### Raw API escape hatch

```bash
shortcut api stories get-story 123 -o json
shortcut api search search-stories --query 'owner:example-user' --detail slim --page_size 5 -o json
```

## Use with AI agents

`shortcut-cli` is designed so an agent can operate it from a terminal session without relying on MCP-specific integrations.

Recommended agent workflow:

1. start with `shortcut --help`
2. inspect the relevant subtree, for example `shortcut search --help`
3. authenticate with `SHORTCUT_API_TOKEN` or `.env`
4. prefer read-only commands unless mutation is explicitly intended
5. request `-o json` when structured output matters
6. fall back to `shortcut api ...` when a curated top-level command is missing

See [`docs/for-ai-agents.md`](docs/for-ai-agents.md) for compact copyable guidance.

## Documentation

- [Docs index](docs/README.md)
- [Usage guide](docs/usage.md)
- [Examples](docs/examples.md)
- [For AI agents](docs/for-ai-agents.md)
- [Release process](docs/releasing.md)
- [Publish checklist](docs/publish-checklist.md)
- [GitHub metadata guidance](docs/github-metadata.md)

## Development

```bash
make build
go test ./...
make dist VERSION=v1.0.0 COMMIT=$(git rev-parse HEAD)
```

Project layout:

- `cmd/shortcut/main.go` — CLI entrypoint
- `internal/cli/` — curated command UX and runtime
- `internal/gen/shortcutv3/` — generated API client
- `openapi/shortcut.openapi.json` — vendored Shortcut OpenAPI source

Do not hand-edit generated files under `internal/gen/shortcutv3/`.

## Release model

Pushing a version tag triggers the release workflow, which:

- runs `go test ./...`
- builds archives for Linux/macOS on amd64/arm64
- generates a checksum manifest
- creates or updates a draft GitHub Release for that tag
- uploads archives, checksums, and `install.sh` to the draft release
