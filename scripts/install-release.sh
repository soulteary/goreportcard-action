#!/usr/bin/env bash
#
# Installs the goreportcard-cli and its linter tools by downloading the
# pre-built binaries from a GitHub Release. Falls back to compiling from the
# module sources (linters pinned via go.mod `tool` directives) when a matching
# release asset cannot be found (or when INPUT_VERSION=source is requested).
#
# Environment:
#   INPUT_VERSION  Release tag to download (e.g. v1.2.3), "latest", or "source".
#   GITHUB_REPOSITORY  owner/repo of this action (set automatically on CI).
#   GH_TOKEN / GITHUB_TOKEN  optional, used to raise GitHub API rate limits.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

VERSION="${INPUT_VERSION:-latest}"
REPO="${GOREPORTCARD_REPO:-gojp/goreportcard}"
BIN_DIR="${RUNNER_TEMP:-/tmp}/goreportcard-bin"

log() { echo "[goreportcard] $*"; }

install_from_source() {
  log "Installing from module sources..."
  "${SCRIPT_DIR}/install-cli.sh"
}

if [ "${VERSION}" = "source" ]; then
  install_from_source
  exit 0
fi

# Map the runner OS/arch to the release archive naming.
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${os}" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  msys*|mingw*|cygwin*) os="windows" ;;
esac

arch="$(uname -m)"
case "${arch}" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
esac

# Resolve the concrete tag when "latest" is requested.
api_base="https://api.github.com/repos/${REPO}/releases"
auth_header=()
token="${GH_TOKEN:-${GITHUB_TOKEN:-}}"
if [ -n "${token}" ]; then
  auth_header=(-H "Authorization: Bearer ${token}")
fi

# Safe expansion of a possibly-empty array under `set -u` (older bash).
curl_auth() {
  if [ "${#auth_header[@]}" -gt 0 ]; then
    curl "${auth_header[@]}" "$@"
  else
    curl "$@"
  fi
}

if [ "${VERSION}" = "latest" ]; then
  tag="$(curl_auth -fsSL "${api_base}/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
else
  tag="${VERSION}"
fi

if [ -z "${tag:-}" ]; then
  log "Could not resolve a release tag (version='${VERSION}'). Falling back to source build."
  install_from_source
  exit 0
fi

ext="tar.gz"
[ "${os}" = "windows" ] && ext="zip"
archive="goreportcard_${tag}_${os}_${arch}.${ext}"
url="https://github.com/${REPO}/releases/download/${tag}/${archive}"

log "Downloading ${archive} from ${tag}..."
mkdir -p "${BIN_DIR}"
if ! curl_auth -fsSL -o "${BIN_DIR}/${archive}" "${url}"; then
  log "Release asset '${archive}' not found. Falling back to source build."
  install_from_source
  exit 0
fi

# Verify the checksum when the release ships checksums.txt.
if curl_auth -fsSL -o "${BIN_DIR}/checksums.txt" \
    "https://github.com/${REPO}/releases/download/${tag}/checksums.txt" 2>/dev/null; then
  log "Verifying checksum..."
  expected="$(grep " ${archive}\$" "${BIN_DIR}/checksums.txt" | awk '{print $1}')"
  if [ -n "${expected}" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "${BIN_DIR}/${archive}" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "${BIN_DIR}/${archive}" | awk '{print $1}')"
    fi
    if [ "${expected}" != "${actual}" ]; then
      log "Checksum mismatch for ${archive} (expected ${expected}, got ${actual})."
      exit 1
    fi
    log "Checksum OK."
  fi
fi

log "Extracting..."
case "${ext}" in
  tar.gz) tar -xzf "${BIN_DIR}/${archive}" -C "${BIN_DIR}" ;;
  zip) unzip -oq "${BIN_DIR}/${archive}" -d "${BIN_DIR}" ;;
esac

# The archive extracts into a versioned directory; move binaries to BIN_DIR root.
extracted="${BIN_DIR}/goreportcard_${tag}_${os}_${arch}"
if [ -d "${extracted}" ]; then
  find "${extracted}" -maxdepth 1 -type f -exec chmod +x {} \; -exec mv {} "${BIN_DIR}/" \;
  rm -rf "${extracted}"
fi

echo "${BIN_DIR}" >> "${GITHUB_PATH:-/dev/stdout}"
log "Installed pre-built binaries (${tag}) to ${BIN_DIR}"
