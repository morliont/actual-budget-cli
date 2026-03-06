# Security Policy

## Supported versions

This project currently supports security fixes on the latest `main` branch and the latest tagged release.

## Reporting a vulnerability

Please do **not** open a public issue for potential vulnerabilities.

Report privately to the maintainers by opening a GitHub Security Advisory draft:

- Go to the repository Security tab
- Click **Report a vulnerability**
- Include impact, reproduction details, and any suggested fix

If Security Advisories are unavailable, open an issue titled `SECURITY: private report requested` with no exploit details and maintainers will provide a private channel.

## Response targets

- Initial acknowledgment: within 72 hours
- Triage/update: within 7 days
- Fix timeline: depends on severity and complexity

## Release integrity and provenance

Tagged releases are produced via GitHub Actions with:

- GoReleaser-generated archives + checksums
- Sigstore cosign **keyless** signature for `checksums.txt`
- GitHub build provenance attestation for `checksums.txt`

Consumers should verify both signature and provenance before use. See:

- [README release verification section](./README.md#verify-release-integrity--provenance)
- [RELEASE.md](./RELEASE.md)
