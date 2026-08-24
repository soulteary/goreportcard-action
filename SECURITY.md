# Security Policy

## Supported versions

Security fixes are applied to the latest `v1` release. Pin the action to a full
commit SHA in production and update it via Dependabot or Renovate.

## Reporting a vulnerability

Please report suspected vulnerabilities privately via GitHub Security Advisories
("Report a vulnerability" on the repository's Security tab) rather than opening a
public issue. Include reproduction steps and the affected version or commit.

## Security model

This action grades a Go project and writes a badge/report back into the
repository. Its trust boundary is small and auditable:

- **Least privilege.** CI and analysis workflows run with `contents: read`.
  Only the badge write-back workflow requests `contents: write`.
- **Pinned actions.** Every third-party GitHub Action is pinned to a full commit
  SHA with a trailing version comment, not a floating tag.
- **No shell injection.** External linters (`gocyclo`, `ineffassign`,
  `misspell`, …) are invoked with explicit process arguments via `exec.Command`,
  never routed through `eval`, `bash -c`, or unquoted shell expansion.
- **Scoped Git writes.** Write-back only ever `git add`s the generated badge and
  report files; it never uses `git add .` / `git add -A`.
- **Release integrity.** Prebuilt binaries are verified against `checksums.txt`
  (SHA256) before use; a failed download or checksum falls back to a source
  build from the pinned action checkout. Each release archive, `checksums.txt`,
  and the SBOM are additionally signed with
  [cosign](https://docs.sigstore.dev/) keyless signing (Sigstore transparency
  log, no long-lived keys), ship a CycloneDX SBOM, and carry an SLSA
  build-provenance attestation. Verify an archive with:

  ```bash
  cosign verify-blob \
    --certificate goreportcard_<version>_<os>_<arch>.tar.gz.pem \
    --signature  goreportcard_<version>_<os>_<arch>.tar.gz.sig \
    --certificate-identity-regexp 'https://github.com/soulteary/goreportcard-action/.+' \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    goreportcard_<version>_<os>_<arch>.tar.gz

  gh attestation verify goreportcard_<version>_<os>_<arch>.tar.gz \
    --repo soulteary/goreportcard-action
  ```
