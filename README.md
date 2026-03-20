# shortcut-cli

`shortcut-cli` is a standalone Go CLI for the official Shortcut REST API.

It gives you:
- a curated top-level UX for common Shortcut workflows
- a full raw `shortcut api ...` surface generated from the vendored OpenAPI spec
- concise text output by default and stable JSON output for automation
- built-in docs and help without requiring MCP

## Install

### Quick install from GitHub Releases

Install the latest release:

```bash
tmp="$(mktemp)" && \
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/latest/download/install.sh -o "$tmp" && \
sh "$tmp" && \
rm -f "$tmp"
```

Install a specific version:

```bash
tmp="$(mktemp)" && \
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/download/v1.0.0/install.sh -o "$tmp" && \
sh "$tmp" --version v1.0.0 && \
rm -f "$tmp"
```

The installer:
- detects `linux` or `darwin`
- detects `amd64` or `arm64`
- downloads the matching archive from GitHub Releases
- verifies its SHA256 checksum
- installs into a user-writable directory already in your `PATH` when possible

Supported release artifacts:
- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

### Install from source with Go

```bash
go install github.com/nazar256/shortcut-cli/cmd/shortcut@latest
```

## Quick start

Check the installed binary:

```bash
shortcut version
shortcut docs summary
shortcut --help
```

Configure your Shortcut token:

```bash
export SHORTCUT_API_TOKEN="your_token"
```

Then try:

```bash
shortcut me
shortcut workflows list
shortcut search stories --query 'owner:ynazarenko is:started'
```

## Common commands

```bash
shortcut me
shortcut stories get 123
shortcut stories get 123 --with-comments
shortcut epics list
shortcut workflows get 500131231
shortcut search syntax
shortcut api stories get-story 123 --output json
```

## Documentation

- Usage guide: [`docs/usage.md`](docs/usage.md)
- Release process: [`docs/releasing.md`](docs/releasing.md)

## Releases

Pushing a version tag triggers the release workflow, which:
- runs `go test ./...`
- builds archives for Linux/macOS on amd64/arm64
- generates a release checksum manifest
- creates or updates a draft GitHub release for that tag
- uploads the archives, checksums, and installer script to that draft release
