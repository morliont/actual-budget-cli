---
name: linear-orchestration-v1
owner: rudy
version: 1.0
tracker: linear
roles:
  orchestrator: Rudy (main assistant)
  implementer: Francois (coding agent)
active_states:
  - Todo
  - In Progress
  - In Review
terminal_states:
  - Done
  - Canceled
branch_pattern: "feat/<linear-key>-<short-slug>"
workspace_pattern: "../actual-budget-cli-<linear-key>"
checks_required:
  - make lint
  - make test
  - make build
retry_policy:
  max_attempts: 3
  backoff: "2m, 10m, then manual intervention"
dispatch_policy: "Rudy dispatches one Linear ticket at a time to Francois unless explicitly batching"
---

# Symphony-style workflow (pragmatic v1)

This repository uses a lightweight orchestration model:

- **Rudy/main** = orchestration and tracking
- **Francois** = implementation and validation
- **Linear** = source of truth for work state

## Prompt template (Rudy → Francois)

Use this template when dispatching a ticket:

```text
[Ticket]
Linear: <KEY> (<URL>)
Title: <TITLE>
State: <STATE>
Priority: <PRIORITY>

[Outcome]
- <expected user-visible result>

[Scope]
In:
- <item>
Out:
- <item>

[Implementation constraints]
- Keep changes minimal and reversible
- No fake integrations; document required env vars

[Validation]
Run and pass:
- make lint
- make test
- make build

[Handoff]
Return:
- summary
- changed files
- risks/follow-ups
- commit hash
```

## State model

### Active states
- `Todo`
- `In Progress`
- `In Review`

### Terminal states
- `Done`
- `Canceled`

## Dispatch policy

1. Rudy selects next Linear ticket in `Todo`.
2. Rudy sends one concrete brief to Francois.
3. Francois executes in a per-ticket branch/workspace.
4. Rudy updates Linear state:
   - `In Progress` at dispatch
   - `In Review` when PR/patch is ready
   - `Done` after merge and successful verification

## Retries / backoff

For failed checks or flaky infra:

- Attempt 1: immediate fix/retry
- Attempt 2: retry after ~2 minutes
- Attempt 3: retry after ~10 minutes with narrowed scope/log capture
- Still failing: hand back to Rudy with blockers and recommended action

## Branch and workspace convention

For ticket `ABC-123` with short slug `improve-login`:

- Branch: `feat/ABC-123-improve-login`
- Optional isolated workspace clone: `../actual-budget-cli-ABC-123`

If no isolated workspace is used, branch locally in this repo with clean `git status` before and after.

## Done / handoff criteria

A ticket is ready for handoff only when all are true:

- Scope complete for ticket intent
- `make lint` passes
- `make test` passes
- `make build` passes
- Docs updated if behavior changed
- Clear summary + changed files + known follow-ups included

