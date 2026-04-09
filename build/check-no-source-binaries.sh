#!/usr/bin/env bash
set -euo pipefail

if command -v cmd.exe >/dev/null 2>&1; then
  echo "warning: build/check-no-source-binaries.sh is deprecated; delegating to build/check-no-source-binaries.bat" >&2
  exec cmd.exe /c build\\check-no-source-binaries.bat
fi

tracked_binaries="$(
  git ls-files \
    '*.exe' \
    '*.dll' \
    '*.so' \
    '*.dylib'
)"

if [[ -z "${tracked_binaries}" ]]; then
  echo "ok: no tracked binary artifacts (.exe/.dll/.so/.dylib)"
  exit 0
fi

echo "error: tracked binary artifacts are not allowed (.exe/.dll/.so/.dylib):" >&2
echo "${tracked_binaries}" >&2
exit 1
