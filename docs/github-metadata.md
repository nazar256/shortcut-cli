# GitHub metadata guidance

These settings live in the GitHub UI, but the intended values are documented here so they are easy to apply consistently.

## Suggested repository description

Single-binary Go CLI for Shortcut, designed for AI agents and automation when MCP is unavailable.

## Suggested topics

- `shortcut`
- `shortcut-api`
- `cli`
- `golang`
- `automation`
- `ai-agents`
- `terminal`
- `openapi`
- `cobra`

## Suggested social preview direction

Use a clean terminal-focused image that shows:

- the `shortcut --help` command tree
- a JSON example such as `shortcut me -o json`
- a short caption like “Shortcut CLI for agents and automation”

Keep it product-specific and avoid generic AI artwork.

## Suggested first release notes outline

Title:

`shortcut-cli v1.0.0`

Body outline:

1. what the tool is: a standalone Shortcut CLI with curated commands and raw API coverage
2. who it is for: AI agents, automation, and terminal-first Shortcut users
3. install options: installer script, direct release downloads, `go install`
4. key examples:
   - `shortcut --help`
   - `shortcut me`
   - `shortcut search stories --query 'owner:example-user is:started' --detail slim`
   - `shortcut me -o json`
5. supported platforms: Linux/macOS on amd64 and arm64

## Homepage guidance

If you want to set a homepage URL, prefer a stable docs destination once one exists. Until then, leaving the GitHub homepage field empty is better than pointing it somewhere redundant.
