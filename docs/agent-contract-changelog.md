# Agent Contract Changelog

Tracks versioned, planner-relevant changes to `--agent-json` behavior and documented agent command contracts.

## Policy

- Bump the contract version for any breaking schema/field/semantic change used by agents.
- Additive, backward-compatible fields should still be recorded as a minor update.
- Every entry should include date, impact, and migration notes.

## Versions

### v1.1.0 — 2026-03-14

- Added optional `meta.correlationId` in agent JSON envelopes for traceability across multi-step orchestration.
- Added `doctor` command with `--agent-json` readiness payload:
  - `data.ready` boolean
  - `data.checks[]` with `name`, `ok`, and optional `details`
  - `data.summary` with `passed` and `total`

Migration notes:
- Existing parsers for `ok`, `data`, and `error` remain valid.
- `meta` is optional and may be ignored by clients that do not need tracing.
- `doctor` is additive and does not change existing command behavior.

### v1.0.0 — 2026-03-06

- Initial envelope contract in `docs/agent-contract.md`.
