# Skill: accounts-list

## Intent
Fetch all accounts from configured budget.

## Command

```bash
actual-cli --agent-json --non-interactive accounts list
```

## Required input

- Existing valid config at `~/.config/actual-cli/config.json`

## Expected output shape (`--agent-json`)

```json
{
  "ok": true,
  "data": {
    "accounts": [
      { "id": "...", "name": "...", "type": "...", "offbudget": false, "closed": false }
    ]
  }
}
```

## Failure handling

- `AUTH_FAILED`: run `auth-check` again (credentials likely stale)
- `NETWORK_UNREACHABLE` / `TIMEOUT`: retry with backoff; verify server
- `INVALID_INPUT`: uncommon here; inspect invocation/global flags
- `INTERNAL_ERROR`: surface error and stop chain
