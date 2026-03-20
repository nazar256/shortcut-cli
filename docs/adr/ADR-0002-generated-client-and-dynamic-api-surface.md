# ADR-0002: Use generated client/models plus a dynamic full-coverage API command surface

## Status
Accepted

## Context
The Shortcut spec is broad and largely untagged. Handcrafting polished commands for every operation in v1 would delay delivery and risk incomplete coverage.

## Decision
Generate a thin Go client/models layer with `oapi-codegen`, keep it isolated from handwritten code, and implement a handwritten CLI/runtime with a dynamic `api` command surface that covers all operations from the spec. Add curated top-level commands for common workflows on top.

## Consequences
- v1 can achieve full API coverage without manually building every command.
- Help output remains discoverable because command metadata comes from the spec.
- Some advanced write operations will rely on JSON request bodies rather than highly tailored flags.

## Alternatives considered
- Handcraft every endpoint command in v1.
- Use raw HTTP only and skip generated client/models.
- Depend on Shortcut MCP parity instead of the official REST spec.
