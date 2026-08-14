#!/usr/bin/env bash
#
# Cross-compiles the goreportcard-cli together with the linter binaries it
# invokes at runtime (gometalinter, gocyclo, ineffassign, misspell) and packages
# them into per-platform archives suitable for a GitHub Release.
#
# Each archive contains all five binaries in its root, so the action can simply
# download one archive, unpack it, and put its contents on PATH.
#
# Usage:
#   scripts/build-release.sh [VERSION]
#
# Environment:
#   PLATFORMS  Space separated list of "os/arch" targets.
#              Defaults to the common CI runner platforms.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

VERSION="${1:-${VERSION:-dev}}"
DIST_DIR="${ROOT_DIR}/dist"
PLATFORMS="${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64}"

# name -> package path (relative to the module or a module dependency).
build_targets() {
  cat <<'EOF'
goreportcard-cli ./cmd/goreportcard-cli
gometalinter github.com/alecthomas/gometalinter
gocyclo github.com/fzipp/gocyclo/cmd/gocyclo
ineffassign github.com/gordonklaus/ineffassign
misspell github.com/client9/misspell/cmd/misspell
EOF
}

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

for platform in ${PLATFORMS}; do
  os="${platform%/*}"
  arch="${platform#*/}"
  ext=""
  if [ "${os}" = "windows" ]; then
    ext=".exe"
  fi

  stage="${DIST_DIR}/goreportcard_${VERSION}_${os}_${arch}"
  mkdir -p "${stage}"

  echo "==> Building for ${os}/${arch}"
  while read -r name pkg; do
    [ -z "${name}" ] && continue
    echo "    - ${name}"
    CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
      go build -trimpath \
      -ldflags "-s -w" \
      -o "${stage}/${name}${ext}" "${pkg}"
  done < <(build_targets)

  # Archive: zip for windows, tar.gz otherwise.
  base="$(basename "${stage}")"
  if [ "${os}" = "windows" ]; then
    (cd "${DIST_DIR}" && zip -q -r "${base}.zip" "${base}")
    archive="${base}.zip"
  else
    (cd "${DIST_DIR}" && tar -czf "${base}.tar.gz" "${base}")
    archive="${base}.tar.gz"
  fi
  rm -rf "${stage}"
  echo "    packaged ${archive}"
done

# Checksums make downloads verifiable.
(cd "${DIST_DIR}" && shasum -a 256 goreportcard_* > "checksums.txt" 2>/dev/null || sha256sum goreportcard_* > "checksums.txt")

echo "==> Artifacts in ${DIST_DIR}:"
ls -1 "${DIST_DIR}"
