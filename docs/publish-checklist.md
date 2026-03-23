# Publish checklist

Use this checklist before publicly promoting the repository.

## Repository files

- [x] README explains what the project is, who it is for, install paths, discovery flow, and examples.
- [x] Usage, examples, AI-agent guidance, release docs, and metadata guidance exist under `docs/`.
- [x] Installer script and release workflow are documented.
- [x] `CONTRIBUTING.md` and `SECURITY.md` exist.
- [ ] Choose a license and add `LICENSE`.
- [ ] Enable GitHub private vulnerability reporting before public launch, then update `SECURITY.md` if you add another private contact path.

## Release readiness

- [ ] Run `go test ./...`.
- [ ] Run `make build`.
- [ ] Run `make dist VERSION=vX.Y.Z COMMIT=$(git rev-parse HEAD)`.
- [ ] Confirm `dist/` contains four platform archives plus a checksum file.
- [ ] Push the release tag and wait for the GitHub `Release` workflow to finish.
- [ ] Inspect the draft GitHub Release and confirm it contains archives, checksums, and `install.sh`.
- [ ] Publish the release.

## Install verification

- [ ] Verify latest install:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/latest/download/install.sh | sh
# if the installer printed a PATH export command, run it first
shortcut version
shortcut docs summary
```

- [ ] Verify pinned install:

```bash
curl -fsSL https://github.com/nazar256/shortcut-cli/releases/download/vX.Y.Z/install.sh | sh -s -- --version vX.Y.Z
# if the installer printed a PATH export command, run it first
shortcut version
```

## Manual GitHub UI steps

- [ ] Set the repository description, topics, and social preview image using [`github-metadata.md`](github-metadata.md).
- [ ] Enable private vulnerability reporting / security advisories.
- [ ] Review and polish the first public release notes.
