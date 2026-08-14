#!/usr/bin/env bash
#
# Installs the goreportcard CLI along with the external linters it depends on.
# It is meant to be run from within the checked-out copy of this action, so it
# uses the linter versions pinned via go.mod `tool` directives to guarantee
# reproducible versions (no vendor directory required).

set -euo pipefail

# Directory of this script -> repository root of the action.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}"

echo "Installing goreportcard linters (pinned via go.mod tool directives)..."
go install tool

echo "Installing goreportcard-cli..."
go install ./cmd/goreportcard-cli

# Ensure GOPATH/bin (where the tools above land) is on PATH for later steps.
GOBIN="$(go env GOPATH)/bin"
echo "${GOBIN}" >> "${GITHUB_PATH:-/dev/stdout}"
echo "Installed CLI and linters to ${GOBIN}"
