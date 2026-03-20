# Curated Command UX Checklist

This checklist tracks the non-`api` command surface that should stay domain-oriented and hide raw transport details.

## Rules

- No `Method:` / `Path:` / transport-envelope details in curated help.
- Text output should not expose `operation_id`, `method`, `path`, `status`, `response`, or similar raw API wrapper fields.
- JSON output for curated commands should return domain-shaped payloads, not the raw transport envelope.
- Legacy generated names may remain as aliases, but help/examples should use canonical concise names.

## Latest validated renderer fixes

- [x] `stories get` hides comments by default and shows comment bodies only with `--with-comments`
- [x] Curated single-item text output preserves large numeric IDs without scientific notation
- [x] Story state text prefers `completed` / `archived` over `started`
- [x] `stories history` has a readable text summary instead of empty/generic bullets

## Top-level curated commands

- [x] `me`
- [x] `stories`
- [x] `epics`
- [x] `iterations`
- [x] `workflows`
- [x] `search`

## Stories

- [ ] `stories create`
- [ ] `stories create-comment`
- [ ] `stories create-from-template`
- [ ] `stories create-multiple`
- [ ] `stories create-reaction`
- [ ] `stories create-task`
- [ ] `stories delete`
- [ ] `stories delete-comment`
- [ ] `stories delete-multiple`
- [ ] `stories delete-reaction`
- [ ] `stories delete-task`
- [ ] `stories get`
- [ ] `stories get-comment`
- [ ] `stories get-task`
- [ ] `stories history`
- [ ] `stories list-comment`
- [ ] `stories list-sub-tasks`
- [ ] `stories query`
- [ ] `stories unlink-comment-thread-from-slack`
- [ ] `stories update`
- [ ] `stories update-comment`
- [ ] `stories update-multiple`
- [ ] `stories update-task`

## Epics

- [ ] `epics create`
- [ ] `epics create-comment`
- [ ] `epics create-comment-comment`
- [ ] `epics create-health`
- [ ] `epics delete`
- [ ] `epics delete-comment`
- [ ] `epics get`
- [ ] `epics get-comment`
- [ ] `epics get-health`
- [ ] `epics list`
- [ ] `epics list-comments`
- [ ] `epics list-documents`
- [ ] `epics list-healths`
- [ ] `epics list-paginated`
- [ ] `epics list-stories`
- [ ] `epics unlink-productboard-from`
- [ ] `epics update`
- [ ] `epics update-comment`

## Iterations

- [ ] `iterations create`
- [ ] `iterations delete`
- [ ] `iterations disable`
- [ ] `iterations enable`
- [ ] `iterations get`
- [ ] `iterations list`
- [ ] `iterations list-stories`
- [ ] `iterations update`

## Workflows

- [ ] `workflows get`
- [ ] `workflows list`

## Search

- [ ] `search all`
- [ ] `search documents`
- [ ] `search epics`
- [ ] `search iterations`
- [ ] `search milestones`
- [ ] `search objectives`
- [ ] `search stories`
- [ ] `search syntax`
