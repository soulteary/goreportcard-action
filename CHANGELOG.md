# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `.golangci.yml` lint config (explicit allow-list) and a `golangci-lint` job in
  CI, pinned to a specific runner version.
- Unit tests for `GradeFromPercentage`, `GradeColor`, `textWidth`,
  `GenerateBadgeSVG`, `GenerateReportMarkdown`, and the CLI (text/JSON/artifact/
  threshold/error paths), raising statement coverage from ~51% to ~76%.
- Coverage collection with a project gate (`>= 70%`) and a patch-coverage gate
  (`scripts/check-patch-coverage.sh`, new lines `>= 90%`) wired into CI on pull
  requests, with a `--selftest` mode CI verifies.
- Supply-chain hardening: a `govulncheck` CI job, a CodeQL analysis workflow, and
  release-time cosign keyless signing, a CycloneDX SBOM, and SLSA build
  provenance attestations for every archive and `checksums.txt`.
- Grouped Dependabot updates for `gomod` and `github-actions`.
- Community files: `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`,
  `CODEOWNERS`, `.editorconfig`, and issue/PR templates.

### Changed
- Renamed the CI workflow `go.yml` to `ci.yml` and expanded it into
  lint/test-matrix/vuln/patch-coverage jobs.
- Pinned all third-party GitHub Actions to full commit SHAs (with a trailing
  version comment) instead of floating major tags.
- Refactored the CLI so its pipeline is a testable `run(options, io.Writer)`.

### Fixed
- Corrected README badge links that mixed the old
  `private-action-goreportcard` slug with `goreportcard-action`.
- Replaced non-existent `actions/checkout@v7` / `actions/setup-go@v7` references
  that would have failed CI.
- Cleaned up stale `.github/FUNDING.yml` placeholder entries.

[Unreleased]: https://github.com/soulteary/goreportcard-action/commits/main
