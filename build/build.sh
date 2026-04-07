#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT}/dist"
mkdir -p "${OUT_DIR}"

VERSION="${VERSION:-v0.1.0}"
COMMIT="${COMMIT:-$(git -C "${ROOT}" rev-parse --short=7 HEAD 2>/dev/null || echo unknown)}"
SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH:-$(date +%s)}"
BUILD_DATE="$(date -u -d "@${SOURCE_DATE_EPOCH}" +%Y-%m-%dT%H:%M:%SZ)"

LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}"

go build -trimpath -buildvcs=false -ldflags "${LDFLAGS}" -o "${OUT_DIR}/goahk" "${ROOT}/cmd/goahk"
