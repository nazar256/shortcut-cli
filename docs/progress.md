# Progress Log

## 2026-03-20
- Initialized the repository as a new standalone project.
- Confirmed repo context is effectively empty except for a local `.env` file.
- Confirmed Shortcut MCP access works for current-member lookup, but MCP coverage is incomplete for the full API goal.
- Gathered architecture guidance: vendored OpenAPI spec, generated client/models, handwritten CLI/runtime, and a full-coverage `api` command surface.
- Decided to ignore normal Shortcut-story branch conventions for this standalone pet project per user instruction.
- Vendored the official Shortcut OpenAPI JSON into `openapi/shortcut.openapi.json` and embedded it for runtime docs/help.
- Replaced placeholder generation with a reproducible `oapi-codegen` pipeline in `internal/gen/shortcutv3`.
- Added explicit env-based config loading, current-directory `.env` support without overriding env vars, and a request editor for generated client auth.
- Reworked the CLI toward a real Cobra command tree with `me`, `api`, `docs`, `version`, and alias entrypoints.
- Wired the dynamic `api` command surface to live HTTP execution for spec-derived operations, including read-only verification against real Shortcut endpoints.
- Added `docs/usage.md` and `.env.example`.
- Polished help text and examples in Cobra commands.
- Replaced alias-style shortcuts with real top-level resource command groups for `stories`, `epics`, `iterations`, `workflows`, and `search`.
- Verified top-level resource commands and dynamic `api` commands both work against live read-only endpoints.
- Fixed reviewer findings around base URL handling, offline docs/version behavior, HTTP error handling, and `$ref` request-body metadata resolution.
- Tightened HTTP success handling to strict 2xx only and updated help rendering so examples appear in `--help`, including top-level resource commands.
- User feedback highlighted that the current help is still too generated-feeling, especially around `search`; next step is a curated search UX with CLI-native syntax guidance and clearer descriptions.
- Reworked `search` into a curated top-level workflow with `stories`, `epics`, `documents`, and `syntax` subcommands, plus examples and query guidance directly in help output.
- Follow-up review found two remaining search-help issues, which were then fixed: the parent `search` help now tells users to choose a scope, and `search documents` now documents structured flags instead of incorrectly describing a generic `--query` flow.
- Search output is now less HTTP-shaped for curated search commands: JSON returns result-focused payloads, text output is summarized, and `--limit` trims noisy result lists.
- Curated resource commands now use shorter context-aware names such as `stories get` and `epics list`, while raw `shortcut api ...` keeps the full generated names as aliases/source-of-truth.
- Verified concise naming across `stories`, `epics`, `iterations`, and `workflows`; legacy generated names like `get-story` still work as aliases under curated commands.
- Added `docs/curated-ux-checklist.md` as an explicit curated-command audit list so every non-`api` endpoint can be reviewed systematically for help/output abstraction.
- Curated single-resource text output now favors readable domain fields over truncated JSON-like blobs; story/epic content is shown in full text mode and extra summary fields are being expanded.
- Continued the curated sweep: `stories get`, `me`, and the generic single-resource renderer now surface domain fields (description, type, estimate, labels, dates, role, workflow states) instead of raw transport envelopes.
- Follow-up walkthrough identified remaining polish issues (scientific-notation IDs, raw-ish `me`, and inconsistent list limiting), and those fixes are being folded into the curated renderer and command builders.
- Added `--with-comments` for curated story/epic `get` commands so long detail views stay readable by default while still allowing comment expansion on demand.
