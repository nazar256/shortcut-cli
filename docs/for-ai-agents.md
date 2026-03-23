# For AI agents

`shortcut` is useful when an agent needs Shortcut access from a normal shell without depending on MCP-specific tooling.

The CLI exposes both read and write operations. Agents should deliberately choose read-only commands unless mutation is part of the requested task.

## Recommended agent workflow

1. Discover the command tree through help:

```bash
shortcut --help
shortcut search --help
shortcut stories --help
shortcut api --help
```

2. Authenticate with an environment variable or `.env` file:

```bash
export SHORTCUT_API_TOKEN="your_token"
```

The CLI also auto-loads `./.env` and `~/.env` unless `--env-file` or `--no-env-file` is used.

3. Prefer JSON when another tool or agent will parse the output:

```bash
shortcut me -o json
shortcut workflows list -o json
shortcut search stories --query 'type:feature' --detail slim -o json
```

4. Use curated commands first, then fall back to the raw API surface when needed:

```bash
shortcut search stories --query 'owner:example-user is:started' --detail slim
shortcut api stories get-story 123 -o json
```

## Practical notes

- `shortcut --help`, `shortcut version`, and `shortcut docs summary` work without API credentials.
- `-o json` is the machine-readable mode.
- `shortcut docs summary` and `shortcut api --help` are useful when an agent needs to inspect the available surface before making a call.
- Search commands expect Shortcut search syntax such as `owner:example-user is:started`; use `shortcut search syntax` for examples.
- Source installs via `go install ...@latest` usually report `dev`/`unknown` version metadata; tagged releases and ldflag-based builds include full version info.

## Copyable examples

```bash
shortcut docs summary -o json
shortcut me -o json
shortcut workflows list -o json
shortcut search stories --query 'label:"bug"' --detail slim -o json
shortcut api search search-stories --query 'type:bug' --detail slim --page_size 5 -o json
```
