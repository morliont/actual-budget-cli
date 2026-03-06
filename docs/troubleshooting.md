# Troubleshooting

## `request timed out after ...`

Increase bridge timeout:

```bash
export ACTUAL_CLI_BRIDGE_TIMEOUT=60s
```

## `network error while contacting Actual server`

- Verify `--server` URL and port
- Ensure server is reachable
- Check VPN/firewall/proxy constraints

## `bridge runtime unavailable`

Install Node.js 20+ and ensure `node` is in `PATH`.

## TLS / self-signed cert issues

Set `NODE_EXTRA_CA_CERTS` to your CA bundle.

## Non-interactive login errors

If running in CI/non-TTY, pass explicit flags:

```bash
actual-cli init --server <URL> --budget <SYNC_ID> --password "$ACTUAL_PASSWORD"
```
