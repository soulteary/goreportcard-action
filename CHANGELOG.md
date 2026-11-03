# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.5] - 2026-11-03

### Fixed
- Capitalize report badge labels and refresh the README badge example.

### Added
- The release workflow can now publish from a manual run (`workflow_dispatch`
  with `publish: true`). Previously the publish step was gated on
  `github.event_name == 'push'`, so a manual run built, signed and attested the
  archives and then discarded them — there was no way to recover a tag-push run
  that failed or never started, short of deleting and re-pushing an already
  released tag. A manual publish refuses a `version` that is not an existing
  tag, and checks out that tag rather than `main`, so the archives are built
  from the released commit.

  Build provenance is skipped when a recovery run's built tree is not
  `GITHUB_SHA`. `actions/attest-build-provenance` takes the source commit for
  its SLSA predicate from the event context and has no input to override it, so
  a recovery dispatched from `main` for an older tag would attest a commit the
  archives were not built from; a wrong source claim is worse than none. The run
  logs a warning saying so. Dispatching the recovery from the tag itself keeps
  the context and the tree in agreement and the attestation is emitted as usual
  — that is the preferred path for any tag whose `release.yml` already carries
  the `publish` input. Tag-push releases are unaffected and always attested.
  `checksums.txt`, the SBOM and the cosign bundles are computed over the actual
  bytes and are emitted either way.

## [1.1.2] - 2026-09-13

### Fixed
- Release signing has never worked. `sigstore/cosign-installer` v4.1.2 installs
  cosign v3.0.6 by default, and cosign v3 defaults to the new bundle format,
  under which `--output-signature` and `--output-certificate` are ignored. With
  no `--bundle` path given, signing failed with
  `create bundle file: open : no such file or directory`, `set -e` stopped the
  job, and the steps that upload the archives never ran — which is why neither
  v1.1.0 nor v1.1.1 carries a single release archive or checksum. Signing now
  writes `<artifact>.cosign.bundle`, which carries both the signature and the
  certificate.
- Made the cosign version explicit via `cosign-release`. It is set to the
  installer's current default, so the binary we run is unchanged; the point is
  that a future installer bump can no longer move cosign without the diff
  saying so. The installer action was pinned by SHA while the version of the
  tool it installs was left implicit.
- The Markdown report embedded a generated-at timestamp, so it was never
  byte-identical between runs. That defeated the "unchanged, skip the commit"
  guard in `commit-badge.sh`: every push to a consumer's default branch produced
  a one-line diff and a write-back commit, even when the grade and every metric
  were identical. In this repository 18 of the last 100 commits on `main` were
  such no-op badge commits, and the badge SVG itself never changed once. The
  footer keeps its attribution link but no longer carries a date, and a test
  asserts the report stays byte-identical across runs.
- Made the publish step idempotent. Creating a Release through the UI or
  `gh release create` also creates the tag that triggers this workflow, so the
  Release is often already present by the time the workflow reaches it and a
  bare `gh release create` failed with "already exists". It now uploads to an
  existing Release instead.

### Changed
- `SECURITY.md`'s verification command now uses `--bundle` instead of
  `--certificate` / `--signature`, matching what the release actually ships.
  The previous command could never have worked: no release has ever carried a
  `.sig` or `.pem` file.

## [1.1.1] - 2026-09-13

### Security
- Pinned `actions/setup-go` in `action.yml` to a commit SHA
  (`b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`, v7.0.0). This was the only
  `uses:` reference in the repository that executes on a *consumer's* runner
  rather than our own, and the only one still on a floating tag — which
  contradicted the guarantee already given in `SECURITY.md`. Consumers run the
  same `setup-go` code as before; it can no longer change without a commit here.

### Changed
- Narrowed the release workflow's tag trigger from `v*` to
  `v[0-9]+.[0-9]+.[0-9]+`. `v*` also matched the floating `v1` / `v1.0`
  aliases, so every manual move of an alias re-fired the full five-platform
  build, SBOM, signing and attestation run, and then called
  `gh release create` against the alias itself. Real releases are unaffected;
  prereleases such as `v1.2.0-rc.1` now require an explicit channel.
- Bumped `anchore/sbom-action` via the grouped `github-actions` Dependabot
  update.

## [1.1.0] - 2026-08-25

### Added
- `.golangci.yml` lint config (explicit allow-list) and a `golangci-lint` job in
  CI, pinned to a specific runner version.
- Unit tests for `GradeFromPercentage`, `GradeColor`, `textWidth`,
  `GenerateBadgeSVG`, `GenerateReportMarkdown`, and the CLI (text/JSON/artifact/
  threshold/error paths), raising statement coverage from ~51% to ~76%.
- Coverage collection with a project gate (`>= 70%`) and a patch-coverage gate
  (`scripts/check-patch-coverage.sh`, new lines `>= 90%`) wired into CI on pull
  requests, with a `--selftest` mode CI verifies.
- Supply-chain hardening: a `govulncheck` CI job, release-time cosign keyless
  signing, a CycloneDX SBOM, and SLSA build provenance attestations for every
  archive and `checksums.txt`.
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
- Replaced the deprecated `gometalinter` driver with native Go tooling
  (`gofmt`, `go vet`, `gocyclo`, `ineffassign`, `misspell`) so grading works on
  modern toolchains.
- Normalized path separators so grading and file URLs work correctly on
  Windows.
- Guarded against out-of-range parts in `goPkgInToGitHub` to avoid a panic on
  malformed `gopkg.in` import paths.
- Corrected README badge links that mixed the old
  `private-action-goreportcard` slug with `goreportcard-action`.
- Replaced non-existent `actions/checkout@v7` / `actions/setup-go@v7` references
  that would have failed CI.
- Cleaned up stale `.github/FUNDING.yml` placeholder entries.

[Unreleased]: https://github.com/soulteary/goreportcard-action/compare/v1.1.2...main
[1.1.2]: https://github.com/soulteary/goreportcard-action/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/soulteary/goreportcard-action/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/soulteary/goreportcard-action/compare/v1.0.0...v1.1.0
