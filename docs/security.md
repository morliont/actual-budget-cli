# Security

## Local config and permissions

`actual-cli` stores config at:

- `~/.config/actual-cli/config.json` (mode `0600`)
- data dir: `~/.local/share/actual-cli` (mode `0700`)

The CLI enforces restrictive file permissions on save.

## Secrets

- Do not commit config files or credentials.
- Prefer environment/secret managers in CI.
- Use secure shells and encrypted disks where possible.

## TLS / custom CAs

For self-signed/private CA servers, set:

```bash
export NODE_EXTRA_CA_CERTS=/path/to/ca-bundle.pem
```

## Release integrity

Releases include checksums, Sigstore keyless signature/cert, and GitHub build provenance.
See [release verification](./release-verification.md).
