# ADR-0001: Vendor the official Shortcut OpenAPI spec as the source of truth

## Status
Accepted

## Context
The CLI must cover the full official Shortcut API and remain reproducible. The repository starts empty, and Shortcut MCP only covers part of the API surface.

## Decision
Commit the official Shortcut OpenAPI spec to `openapi/shortcut.openapi.json` and treat it as the canonical source for generated code, command metadata, and embedded help/docs.

## Consequences
- API coverage can track the official spec directly.
- Code generation and command metadata become reproducible.
- Updating the CLI for API changes requires an explicit spec refresh step.

## Alternatives considered
- Fetch the spec dynamically at build/runtime.
- Use Shortcut MCP or an unofficial SDK as the primary contract.
