#!/usr/bin/env bash
set -euo pipefail

if command -v cmd.exe >/dev/null 2>&1; then
  echo "warning: build/check-no-source-binaries.sh is deprecated; delegating to build/check-no-source-binaries.bat" >&2
  exec cmd.exe /c build\\check-no-source-binaries.bat
fi

tracked_exes="$(git ls-files '*.exe')"
if [[ -z "${tracked_exes}" ]]; then
  echo "ok: no tracked .exe artifacts"
  exit 0
fi

echo "error: tracked .exe artifacts are not allowed:" >&2
echo "${tracked_exes}" >&2
exit 1
