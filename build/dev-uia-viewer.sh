#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="${ROOT}/cmd/goahk-uia-viewer"
DIST_DIR="${ROOT}/dist/goahk-uia-viewer"

mkdir -p "${DIST_DIR}"

cd "${APP_DIR}"
wails dev
