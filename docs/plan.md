# Shortcut CLI v1 Plan

## Objective
Build a fully working Go CLI for the official Shortcut REST API that is easy for AI agents to discover and use through command names and built-in help output.

## Business goal
Enable AI-agent and human access to Shortcut without requiring MCP connectivity, while preserving strong discoverability and full API coverage.

## In scope (v1)
- Official Shortcut OpenAPI v3 as the API source of truth.
- Go CLI implemented with Cobra.
- Full API coverage through CLI commands derived from the spec.
- Generated client/models isolated from handwritten code.
- Built-in help and usage examples that expose how to use the CLI without external docs.
- `SHORTCUT_API_TOKEN` auth via environment variables.
- Optional `.env` loading for local development.
- Concise default output plus stable JSON output for automation.
- Read-only real-system verification against the live API.

## Non-goals
- Interactive TUI workflows.
- Non-Go implementations.
- Maintaining compatibility with unofficial third-party Shortcut SDKs.
- Broad mutation-heavy end-to-end testing against the real workspace.
- Perfect handcrafted UX for every single operation in v1.

## Constraints
- New standalone repo with no preexisting codebase.
- Official docs/spec should be preferred over rediscovery.
- Generated artifacts must remain reproducible.
- Secrets must not be committed or copied into repo docs.

## Execution order
1. Initialize repo structure and planning artifacts.
2. Vendor the official OpenAPI spec and set up reproducible code generation.
3. Build config/auth/runtime/output foundations.
4. Implement dynamic full-coverage `api` command surface from the spec.
5. Add curated top-level commands for key workflows.
6. Add embedded docs/help and usage guidance.
7. Validate with tests, build, help snapshots, and live read-only checks.

## Current refinement focus
- Replace awkward generated-feeling help copy with direct user-oriented descriptions.
- Redesign `search` as a curated workflow command instead of exposing raw generated names like `search search-stories` as the primary path.
- Make search query usage and syntax discoverable from the CLI itself, while keeping structured document search help accurate.
- Keep `shortcut api ...` as the stable full-coverage raw surface while improving top-level human-facing commands.

## Testing approach
- Unit tests for config precedence, spec parsing, command metadata, and output formatting.
- CLI tests for help output and argument handling.
- Reproducibility check for generated code.
- Read-only live verification using the existing local token via env/`.env`.
