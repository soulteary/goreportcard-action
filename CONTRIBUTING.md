# Contributing

Thanks for your interest in improving **goreportcard-action**.

## Development setup

Requires Go (see [`go.mod`](go.mod) for the version) and `make`. The CLI drives
external linters, so install those once with `make install`.

```bash
git clone https://github.com/soulteary/goreportcard-action
cd goreportcard-action
make install    # install goreportcard-cli and the linters it invokes
make all        # lint (fmt + vet + misspell) + build + test
```

## Common tasks

| Command | What it does |
| --- | --- |
| `make build` | Build all packages. |
| `make test` | Run tests with coverage. |
| `make lint` | Run `gofmt`, `go vet`, and `misspell`. |
| `make badge` | Regenerate `.github/goreportcard.svg` locally. |
| `golangci-lint run ./check/... ./cmd/...` | Run the configured linters. |
| `go test -covermode=atomic -coverprofile=coverage.out ./check/... ./cmd/...` | Collect coverage. |
| `bash scripts/check-patch-coverage.sh --selftest` | Verify the patch-coverage gate itself. |

Install `golangci-lint` locally with:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Pull requests

- Add or update tests for behavior changes. CI enforces project coverage
  `>= 70%` and patch coverage `>= 90%` on newly added lines.
- Run `make all` and `golangci-lint run ./check/... ./cmd/...` before pushing.
- Keep [`README.md`](README.md), [`action.yml`](action.yml), and the CLI flags
  in sync when you add or change inputs/outputs.
- Update the [`CHANGELOG.md`](CHANGELOG.md) `Unreleased` section for user-facing
  changes.

## Reporting security issues

Please follow [`SECURITY.md`](SECURITY.md) rather than opening a public issue.

## Code of conduct

This project adheres to the [Contributor Covenant](CODE_OF_CONDUCT.md).
