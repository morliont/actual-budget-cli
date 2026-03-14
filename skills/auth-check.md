# Skill: auth-check

## Intent
Validate saved credentials/connectivity without changing local config.

## Command

```bash
# Validate current saved config/session (no writes)
actual-cli --agent-json --non-interactive auth check
```

## Required input

- Existing valid config at `~/.config/actual-cli/config.json`

## Expected output shape (`--agent-json`)

Success:

```json
{ "ok": true, "data": { "authenticated": true, "message": "Credentials are valid" } }
```

Failure:

```json
{ "ok": false, "error": { "code": "INVALID_INPUT|AUTH_FAILED|...", "message": "...", "retryable": false } }
```

## Failure handling

- `INVALID_INPUT`: fix flags/value format; do not retry unchanged
- `AUTH_FAILED`: verify password/server/budget ID
- `NETWORK_UNREACHABLE` or `TIMEOUT`: retry with backoff; check server reachability
- `INTERNAL_ERROR`: capture stderr + context and escalate
