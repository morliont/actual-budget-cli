# CLAUDE.md — Agent Entrypoint

Start here for agent-oriented work in this repository.

## Primary references

1. [AGENTS.md](./AGENTS.md) — operator quickstart, command ordering, failure semantics
2. [docs/agent-contract.md](./docs/agent-contract.md) — canonical `--agent-json` envelope contract
3. [docs/capability-map.md](./docs/capability-map.md) — intent → command routing map
4. [`.claude/skills/`](./.claude/skills/) — capability-specific skills

## Operating rules (concise)

- Prefer deterministic execution: `actual-cli --agent-json --non-interactive ...`
- Parse only the envelope contract (`ok`, `data`, `error`, `meta`)
- Treat `error.retryable=true` as bounded retry with backoff; otherwise stop and remediate
- Do not log or persist secrets (passwords, tokens, local config)
- Keep changes minimal and compatible with documented command/JSON contracts
