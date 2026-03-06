# Advanced usage

## First-run command

`actual-cli init` is a TTY-aware wizard for first-run setup. In non-interactive environments, provide required flags:

```bash
actual-cli init --server <URL> --budget <SYNC_ID> --password "$ACTUAL_PASSWORD"
```

## JSON output

Machine-readable output:

```bash
actual-cli accounts list --json
actual-cli transactions list --json
actual-cli budgets summary --json
```

## Transaction filters

```bash
actual-cli transactions list \
  --account <ACCOUNT_ID> \
  --from 2026-01-01 \
  --to 2026-01-31 \
  --limit 50
```

## Bridge timeout

Default is `30s`. Override with:

```bash
export ACTUAL_CLI_BRIDGE_TIMEOUT=45s
```
