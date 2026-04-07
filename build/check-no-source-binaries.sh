#!/usr/bin/env bash
set -euo pipefail

tracked_exes="$(git ls-files '*.exe')"
if [[ -z "${tracked_exes}" ]]; then
  echo "ok: no tracked .exe artifacts"
  exit 0
fi

echo "error: tracked .exe artifacts are not allowed:" >&2
echo "${tracked_exes}" >&2
exit 1
