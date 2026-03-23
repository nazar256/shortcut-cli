# Shortcut CLI Usage Guide

Shortcut CLI exposes the official Shortcut REST API in a way that is easy to discover from `--help`, friendly for automation, and safe to validate against the real API with read-only commands.

## Installation

If the repository already has a published GitHub Release, you can install from the release assets:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/latest/download/install.sh | sh
```

If the installer picks a directory that is not already on `PATH`, it prints the export command to run before you invoke `shortcut` from a new shell.

Install a specific version:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/download/v1.0.0/install.sh | sh -s -- --version v1.0.0
```

If you prefer to inspect the installer before running it, download it first:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/latest/download/install.sh -o ./install-shortcut.sh
sh ./install-shortcut.sh
rm -f ./install-shortcut.sh
```

Or install from source:

```bash
go install github.com/nazar256/shortcut-cli/cmd/shortcut@latest
```

The source install path works even before the first public release exists.

Building from source requires Go `1.25+`.

Note: source installs do not inject release build metadata, so `shortcut version` usually reports `dev` / `unknown` values unless you build with ldflags.

For release details, see [`docs/releasing.md`](releasing.md).

By default, the installer chooses the destination in this order:

1. `--install-dir <dir>` if provided
2. the first writable canonical directory already present in `PATH`, excluding known language-managed bins such as `nvm`, configured/default pnpm and Cargo bins, and Go bin directories
3. `~/.local/bin`
4. `~/bin`

Commands like `shortcut docs ...` and `shortcut version` work offline and do not require a token.

The CLI exposes both read and write operations. If you only need read-only verification, choose commands such as `shortcut me`, `shortcut workflows list`, `shortcut docs summary`, and read-only search/resource lookups.

See also:

- [`examples.md`](examples.md)
- [`for-ai-agents.md`](for-ai-agents.md)

## Configuration

The CLI reads configuration from environment variables and can also load dotenv files.

Default dotenv search order (when no dotenv flags are passed):

1. `./.env`
2. `~/.env`

If neither file exists, the CLI continues without error.

Priority is:

1. explicit CLI flags
2. process environment variables
3. dotenv values
4. built-in defaults

Notes:

- `--env-file <path>` loads only that file and disables automatic search.
- `--no-env-file` disables dotenv loading entirely.
- Process environment variables always win over dotenv values, even when set to an empty string.

Supported variables:

- `SHORTCUT_API_TOKEN` — required API token
- `SHORTCUT_BASE_URL` — optional API base URL, defaults to `https://api.app.shortcut.com`
- `SHORTCUT_TIMEOUT` — optional HTTP timeout duration, defaults to `30s`

Example:

```bash
export SHORTCUT_API_TOKEN="your_token"
export SHORTCUT_TIMEOUT="20s"
```

Or create a local `.env` file:

```env
SHORTCUT_API_TOKEN=your_token
SHORTCUT_TIMEOUT=20s
```

## Global flags

- `-o, --output text|json` — choose concise text or stable JSON output
- `--env-file <path>` — load dotenv values only from the provided file
- `--no-env-file` — skip dotenv loading entirely
- `-h, --help` — show built-in help

## Top-level commands

Main command groups:

- `shortcut me` — get the authenticated member
- `shortcut docs` — inspect embedded OpenAPI-derived docs
- `shortcut version` — print CLI version info
- `shortcut api` — full spec-driven API surface
- `shortcut stories` — story commands
- `shortcut epics` — epic commands
- `shortcut iterations` — iteration commands
- `shortcut workflows` — workflow commands
- `shortcut search` — curated search commands with built-in syntax help
- `shortcut completion` — generate shell completion scripts

## Discovering the CLI

Start from help:

```bash
shortcut --help
shortcut stories --help
shortcut search --help
shortcut search syntax
shortcut api --help
shortcut docs summary
```

Inspect a single operation:

```bash
shortcut docs operation stories get-story
shortcut api stories get-story --help
```

## Common examples

### Member info

```bash
shortcut me
shortcut me --output json
```

### Workflow listing

```bash
shortcut workflows list
shortcut api workflows list-workflows --output json
```

Curated top-level resource commands use concise names, for example:

```bash
shortcut stories get 123
shortcut stories query --body '{"workflow_state_id":500131237}'
shortcut stories get 123 --with-comments
shortcut epics list
shortcut workflows get 500131231
shortcut iterations list
```

### Search stories

```bash
shortcut search stories --query "id:sc-12345"
shortcut search stories --query "owner:example-user is:started" --detail slim
shortcut search stories --query "label:\"ios\" updated:2026-03-01..2026-03-20" --detail slim
shortcut api search search-stories --query "owner:example-user" --detail slim --page_size 5 --output json
```

### Learn search syntax

```bash
shortcut search syntax
shortcut search help
shortcut search stories --help
shortcut search stories --limit 5 --query "owner:example-user is:started"
```

### Get a single resource

```bash
shortcut stories get 123
shortcut epics get 456
shortcut iterations get 789
```

### Send JSON request bodies

```bash
shortcut api stories create-story --body '{"name":"Example","project_id":123}'
shortcut api stories update-story 123 --body-file ./update-story.json
```

### Multipart upload

```bash
shortcut api files upload-files --form story_id=123 --file file0=./attachment.txt
```

## API command model

The `api` command is generated from the vendored OpenAPI spec. Commands are grouped by resource and operations are named from `operationId` when available.

Examples:

```bash
shortcut api categories list-categories
shortcut api stories get-story 123
shortcut api search search-stories --query "type:bug"
shortcut api workflows get-workflow 500131231
```

Path parameters are positional. Query parameters are flags. JSON body inputs use `--body` or `--body-file`. Multipart endpoints use repeated `--form` and `--file` flags.

## Built-in docs

The CLI embeds the vendored OpenAPI spec and can summarize it without internet access:

```bash
shortcut docs summary
shortcut docs operation workflows get-workflow
```

## Read-only real-system verification

These commands are safe read-only checks that were used for live verification:

```bash
shortcut me --output json
shortcut docs summary
shortcut workflows list --output json
shortcut search stories --query "type:feature" --detail slim --page_size 1 --output json
```

## Notes

- The CLI does not require MCP connectivity.
- The official Shortcut REST API spec is the source of truth.
- Generated client code lives separately from handwritten CLI/runtime code.
