# Orchestration guide (Rudy + Francois + Linear)

This is a practical day-to-day flow, not a heavy platform.

## Roles

- **Rudy (main assistant / orchestrator)**
  - picks ticket
  - sets scope
  - dispatches to Francois
  - updates Linear states
- **Francois (coding agent)**
  - implements code/docs
  - runs checks
  - reports concise handoff
- **Linear (tracker)**
  - stores priority/state/history

## Daily flow

1. **Select ticket in Linear** (`Todo`)
2. **Move to `In Progress`** when work starts
3. **Generate brief** (from issue payload or URL)
4. **Dispatch brief** to Francois
5. **Francois implements + validates**
6. **Open PR / provide patch** and move Linear to `In Review`
7. **Merge and verify**, then move to `Done`

## Commands

### Generate a standardized brief

From Linear URL (key extraction only; no API required):

```bash
node scripts/linear-brief.mjs --url "https://linear.app/<team>/issue/ABC-123/some-title"
```

From webhook/export payload JSON:

```bash
node scripts/linear-brief.mjs --payload path/to/linear-issue.json
```

Fetch issue details from Linear API (optional):

```bash
LINEAR_API_KEY=lin_api_xxx \
node scripts/linear-brief.mjs --url "https://linear.app/<team>/issue/ABC-123/some-title" --api
```

Dry run (never calls API):

```bash
node scripts/linear-brief.mjs --url "https://linear.app/<team>/issue/ABC-123/some-title" --dry-run
```

### Required validation for each ticket

```bash
make lint
make test
make build
```

## Failure handling

### Check failures

- Fix obvious issue and rerun once
- If still failing, apply backoff (`~2m`, then `~10m`)
- If still blocked, hand back with:
  - exact failing command
  - shortest reproducible error
  - likely root cause
  - recommended next action

### Scope ambiguity

- Stop and request scope clarification from Rudy
- Do not silently expand scope

### Missing credentials/integrations

- Do not fake integrations
- Document exact env vars needed and provide dry-run path

## Operator runbook (for Thijs)

Use this when driving work quickly tomorrow.

1. In Linear, pick one `Todo` ticket.
2. Generate brief:
   - URL only: `node scripts/linear-brief.mjs --url "<linear-issue-url>"`
   - or payload file: `node scripts/linear-brief.mjs --payload issue.json`
3. Paste brief to Rudy and ask to dispatch to Francois.
4. Wait for handoff summary + commit hash.
5. Ensure checks are green (`make lint && make test && make build`).
6. Review PR, merge, set Linear to `Done`.

Minimal rule: one ticket in flight per agent unless explicitly batching.
