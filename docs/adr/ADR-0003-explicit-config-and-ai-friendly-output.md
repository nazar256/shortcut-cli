# ADR-0003: Use explicit env-based configuration and concise default output

## Status
Accepted

## Context
The CLI is intended for both humans and AI agents. Configuration and output should be predictable and easy to reason about.

## Decision
Support `SHORTCUT_API_TOKEN` via environment variables as the default auth mechanism, allow optional `.env` loading for local development without overriding already-set environment variables, and provide concise text output by default with stable JSON output for automation.

## Consequences
- Configuration behavior is explicit and easy to document.
- AI agents can rely on compact machine-readable output when needed.
- Local development remains convenient without committing secrets.

## Alternatives considered
- Introduce a broader config framework such as Viper.
- Use verbose JSON as the default output mode.
- Support deprecated query-parameter authentication as a primary flow.
