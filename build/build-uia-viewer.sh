#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT}/dist/goahk-uia-viewer"

mkdir -p "${DIST_DIR}"
rm -rf "${DIST_DIR:?}"/*

cd "${ROOT}"
go build -o "${DIST_DIR}/goahk-uia-viewer.exe" ./cmd/goahk-uia-viewer
