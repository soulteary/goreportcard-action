# Go Report Card

[![Go Report Card](./.github/goreportcard.svg)](./.github/goreportcard-report.md)
[![Go](https://github.com/soulteary/private-action-goreportcard/actions/workflows/go.yml/badge.svg)](https://github.com/soulteary/private-action-goreportcard/actions/workflows/go.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)

Generate a quality report card for a Go project and turn it into a
self-contained SVG badge. It runs several checks — `gofmt`, `go vet`,
`gocyclo`, `ineffassign`, `misspell` and a license check — and produces an
overall grade (from `A+` down to `F`).

This project ships two things:

- a **command line tool** (`goreportcard-cli`) for grading a project locally, and
- a **GitHub Action** that grades a repository, renders an SVG badge, and can
  commit the badge back into the repository.

## Quick start

Grade a project on your machine:

```sh
# Install the CLI.
go install github.com/soulteary/goreportcard-action/cmd/goreportcard-cli@latest

# Grade the current directory.
goreportcard-cli
```

> The CLI drives external linters (`gocyclo`, `ineffassign`, `misspell`, …). To
> install those alongside the CLI in one step, clone the repo and run
> `make install` (see [Command line interface](#command-line-interface) below).

Or add it to CI as a GitHub Action that grades your repo and commits a badge:

```yaml
- uses: soulteary/goreportcard-action@v1
  with:
    directory: "."
    output: "goreportcard.svg"
    commit: "true"
```

The sections below cover each entry point in detail.

## Command line interface

```sh
git clone https://github.com/soulteary/goreportcard-action.git
cd goreportcard-action
make install
go install ./cmd/goreportcard-cli
goreportcard-cli
```

```
Grade ........... A+ 96.5%
Files ............... 42
Issues ............... 1
gofmt .............. 100%
go_vet ............. 100%
gocyclo ............ 100%
ineffassign ........ 100%
license ............ 100%
misspell ........... 100%
```

Verbose output lists the offending files:

```sh
goreportcard-cli -v
```

Generate a self-contained SVG badge (no external service required):

```sh
goreportcard-cli -svg .github/goreportcard.svg
```

Generate a Markdown quality report:

```sh
goreportcard-cli -report goreportcard-report.md
```

Useful flags:

| Flag     | Description                                                            |
| -------- | --------------------------------------------------------------------- |
| `-d`     | Root directory of the Go project (default `.`).                       |
| `-v`     | Verbose output.                                                       |
| `-j`     | JSON output (always exits 0).                                         |
| `-t`     | Failure threshold as a score percentage; exits non-zero when below.   |
| `-svg`   | Write a self-contained badge SVG to the given path.                   |
| `-report`| Write a Markdown quality report to the given path.                    |

Common recipes:

```sh
# Grade a project in another directory.
goreportcard-cli -d ./path/to/project

# Fail (exit code 1) when the score drops below 90% — useful as a CI gate.
goreportcard-cli -t 90

# Emit machine-readable JSON (always exits 0) for scripting or dashboards.
goreportcard-cli -j | jq '.GradeFromPercentage, .average'

# Produce the badge and the Markdown report in one run.
goreportcard-cli -svg .github/goreportcard.svg -report .github/goreportcard-report.md
```

## GitHub Action

The repository is itself a GitHub Action. It grades a Go project, generates a
self-contained SVG badge (no dependency on `shields.io`), writes a Markdown
quality report, and can commit both back into your repository.

Add the following workflow to your project at
`.github/workflows/goreportcard.yml`:

```yaml
name: Go Report Card

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: write

jobs:
  report-card:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - id: grc
        uses: soulteary/goreportcard-action@v1
        with:
          directory: "."
          output: ".github/goreportcard.svg"
          commit: "true"
      - run: echo "Grade ${{ steps.grc.outputs.grade }} (${{ steps.grc.outputs.score }}%)"
```

Then reference the committed badge from your README:

```markdown
![Go Report Card](./.github/goreportcard.svg)
```

### Inputs

| Input               | Default                                          | Description                                                                    |
| ------------------- | ------------------------------------------------ | ------------------------------------------------------------------------------ |
| `directory`         | `.`                                              | Root directory of the Go project to grade.                                     |
| `output`            | `.github/goreportcard.svg`                       | Path (relative to the repo) where the badge is written.                        |
| `report`            | `true`                                           | Whether to also generate a Markdown quality report.                            |
| `report_output`     | `.github/goreportcard-report.md`                 | Path (relative to the repo) where the Markdown report is written.              |
| `threshold`         | `0`                                              | Minimum score percentage (0-100). The action fails below it (after the badge/report are committed). `0` never fails. |
| `commit`            | `true`                                           | Whether to commit the generated badge back to the repository.                  |
| `commit_message`    | `chore: update Go Report Card badge [skip ci]`   | Commit message used when committing the badge.                                 |
| `commit_user_name`  | `github-actions[bot]`                            | Git `user.name` used for the commit.                                           |
| `commit_user_email` | `github-actions[bot]@users.noreply.github.com`   | Git `user.email` used for the commit.                                          |
| `version`           | `latest`                                          | Release tag of the pre-built binaries to download (e.g. `v1.2.3`), `latest`, or `source` to compile from the module sources. |

By default the action downloads pre-built binaries (the CLI plus the linter
tools) from this repository's GitHub Releases, which avoids compiling anything
at run time. If a matching release asset cannot be found, it automatically falls
back to compiling from the module sources. Set `version: source` to always
compile.

### Outputs

| Output  | Description                                  |
| ------- | -------------------------------------------- |
| `grade` | The overall letter grade (e.g. `A+`).        |
| `score` | The overall score percentage (e.g. `94.1`).  |
| `badge` | The path to the generated SVG badge.         |
| `report`| The path to the generated Markdown report (empty when disabled). |

A ready-to-copy example lives in [`examples/goreportcard-action.yml`](examples/goreportcard-action.yml).

## Releasing (maintainers)

Pushing a tag that starts with `v` triggers the
[`Release`](.github/workflows/release.yml) workflow, which cross-compiles the
CLI and the linter tools for Linux, macOS and Windows (amd64/arm64) and uploads
per-platform archives plus a `checksums.txt` to a GitHub Release:

```sh
git tag v1.2.3
git push origin v1.2.3
```

You can reproduce the artifacts locally with:

```sh
scripts/build-release.sh v1.2.3   # writes archives to ./dist
```

## Contributing

Contributions are welcome! Feel free to raise an issue or submit a pull request.

## License

The code is licensed under the permissive Apache v2.0 license. [Read this](https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)) for a summary.
