# Release Process

## Prerequisites

- Go 1.22+
- Node.js 20+
- `goreleaser` installed locally for dry-runs
- `cosign` and `gh` CLI for local verification (optional but recommended)

## Tag-based release

Releases are created automatically when pushing a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The `release.yml` workflow will:

1. run `make lint`, `make test`, `make build`
2. execute GoReleaser (`goreleaser release --clean`)
3. sign `dist/checksums.txt` with **cosign keyless** (GitHub OIDC identity)
4. verify the generated cosign signature in CI
5. create a GitHub build provenance attestation for `dist/checksums.txt`
6. upload `checksums.txt.sig` and `checksums.txt.pem` to the GitHub release

## Local dry run

```bash
make goreleaser-check
make goreleaser-dry-run
```

## User verification commands

Replace `vX.Y.Z` and asset name as needed.

```bash
TAG=vX.Y.Z
ASSET=actual-cli_${TAG}_linux_amd64.tar.gz

curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/${ASSET}
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt.sig
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt.pem

cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp '^https://github.com/morliont/actual-budget-cli/.github/workflows/release.yml@refs/tags/v.*$' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt

sha256sum --ignore-missing -c checksums.txt

gh attestation verify checksums.txt \
  --repo morliont/actual-budget-cli
```
