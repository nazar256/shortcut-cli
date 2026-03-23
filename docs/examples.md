# Examples

These examples focus on terminal discovery, read-only workflows, and machine-readable output.

## Discover the command tree

```bash
shortcut --help
shortcut search --help
shortcut stories --help
shortcut docs summary
shortcut api --help
```

Inspect one raw API operation in detail:

```bash
shortcut docs operation stories get-story
shortcut api stories get-story --help
```

## Basic auth and connectivity checks

```bash
export SHORTCUT_API_TOKEN="your_token"
shortcut me
shortcut me -o json
```

If you prefer dotenv files:

```bash
shortcut me
```

Create `.env` manually unless you are already in a repository checkout:

```env
SHORTCUT_API_TOKEN=your_token
SHORTCUT_TIMEOUT=20s
```

## Read-only Shortcut workflows

List workflows:

```bash
shortcut workflows list
shortcut workflows list -o json
```

Fetch a story:

```bash
shortcut stories get 123
shortcut stories get 123 --with-comments
shortcut stories get 123 -o json
```

List epics and iterations:

```bash
shortcut epics list
shortcut iterations list
```

## Search examples

Search active work for a member:

```bash
shortcut search stories --query 'owner:example-user is:started' --detail slim
```

Search by ID or label:

```bash
shortcut search stories --query 'id:sc-12345'
shortcut search stories --query 'label:"ios" updated:2026-03-01..2026-03-20' --detail slim
```

Search across record types:

```bash
shortcut search all --query '"checkout" type:epic'
```

Learn the search syntax from the CLI itself:

```bash
shortcut search syntax
shortcut search stories --help
```

## JSON output for scripts and agents

```bash
shortcut docs summary -o json
shortcut version -o json
shortcut workflows list -o json
shortcut search stories --query 'type:bug' --detail slim -o json
```

## Raw API examples

The curated commands are the preferred UX. Use `shortcut api ...` when you need exact raw API coverage.

```bash
shortcut api stories get-story 123 -o json
shortcut api workflows get-workflow 500131231 -o json
shortcut api search search-stories --query 'owner:example-user' --detail slim --page_size 5 -o json
```
