# Release verification

For release `vX.Y.Z`:

```bash
TAG=vX.Y.Z
ASSET=actual-cli_${TAG}_linux_amd64.tar.gz

curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/${ASSET}
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt.sig
curl -LO https://github.com/morliont/actual-budget-cli/releases/download/${TAG}/checksums.txt.pem
```

## 1) Verify checksum signature (Sigstore keyless)

```bash
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp '^https://github.com/morliont/actual-budget-cli/.github/workflows/release.yml@refs/tags/v.*$' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
```

## 2) Verify asset checksum

```bash
sha256sum --ignore-missing -c checksums.txt
```

## 3) Verify provenance attestation

```bash
gh attestation verify checksums.txt --repo morliont/actual-budget-cli
```
