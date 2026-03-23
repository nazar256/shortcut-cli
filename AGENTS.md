# Repository Guide for Agents and Contributors

## Project shape

- Binary name: `shortcut`
- Module path: `github.com/nazar256/shortcut-cli`
- CLI entrypoint: `cmd/shortcut/main.go`
- Curated command UX lives under `internal/cli/`
- Vendored OpenAPI spec: `openapi/shortcut.openapi.json`
- Generated client: `internal/gen/shortcutv3/`

## Common commands

```bash
go test ./...
make build
make dist VERSION=v1.0.0 COMMIT=$(git rev-parse HEAD)
```

## Release model

- CI runs on pushes to `main` and pull requests.
- Release artifacts are built by `.github/workflows/release.yml`.
- Publishing a GitHub Release triggers binary builds for:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
- Release archives are uploaded back to the GitHub Release together with a SHA256 checksum manifest.

## Installer

- Public installer: `install.sh`
- Default install path policy:
  1. explicit `--install-dir`
  2. first writable canonical directory already in `PATH`, excluding known language-managed bins (for example nvm, configured/default pnpm and cargo bins, Go bin dirs)
  3. `~/.local/bin`
  4. `~/bin`
- Installer verifies archive checksum before extraction.
- For local testing, `SHORTCUT_INSTALL_BASE_URL` may override the release asset base URL.

## Editing guidance

- Keep curated commands domain-oriented; raw transport details belong only under `shortcut api ...`.
- Do not hand-edit generated client files under `internal/gen/shortcutv3/`.
- Prefer focused, reviewable changes over broad refactors.
