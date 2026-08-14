#!/bin/sh
#
# Installs the external linters that the goreportcard checks depend on.
# The tools are pinned via the `tool` directives in go.mod, so they resolve
# to reproducible versions from the module cache (no vendor directory needed).

set -e

go install tool
