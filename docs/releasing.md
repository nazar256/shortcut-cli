# Releasing shortcut-cli

This project builds release binaries through GitHub Actions and attaches them to a draft GitHub Release before publication.

## What the release workflow does

When a version tag like `v1.0.0` is pushed, the workflow:

1. checks out the tagged source
2. runs `go test ./...`
3. builds `shortcut` for:
   - `linux/amd64`
   - `linux/arm64`
   - `darwin/amd64`
   - `darwin/arm64`
4. packages each binary as `tar.gz`
5. generates `shortcut-cli_<tag>_checksums.txt`
6. creates a draft GitHub Release if it does not already exist
7. uploads all artifacts back to that draft release

## Artifact naming

Examples for `v1.0.0`:

- `shortcut-cli_v1.0.0_linux_amd64.tar.gz`
- `shortcut-cli_v1.0.0_linux_arm64.tar.gz`
- `shortcut-cli_v1.0.0_darwin_amd64.tar.gz`
- `shortcut-cli_v1.0.0_darwin_arm64.tar.gz`
- `shortcut-cli_v1.0.0_checksums.txt`

Each archive contains a single binary named `shortcut`.

Each published release also uploads `install.sh` so pinned install commands can use the installer that shipped with that release.

## Local preflight

Before cutting a release, run:

```bash
make test
make dist VERSION=v1.0.0 COMMIT=$(git rev-parse HEAD)
```

Inspect artifacts:

```bash
ls dist/
tar -tzf dist/shortcut-cli_v1.0.0_linux_amd64.tar.gz
```

## Publishing a release

1. Commit and push `main`
2. Create and push a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

3. Wait for the `Release` GitHub Actions workflow to finish
4. Open the draft GitHub Release created for that tag
5. Confirm the attached artifacts look correct
6. Publish the draft release

## Post-release verification

Confirm the release page contains all six assets:

- four platform archives
- one checksum manifest
- `install.sh`

Then verify the public installer:

```bash
tmp="$(mktemp)" && \
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/download/v1.0.0/install.sh -o "$tmp" && \
sh "$tmp" --version v1.0.0 && \
rm -f "$tmp"
shortcut version
shortcut docs summary
```

If your install directory is not already on `PATH`, the installer prints the export command to use.
