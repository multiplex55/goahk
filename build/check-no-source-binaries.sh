#!/usr/bin/env bash
set -euo pipefail

if command -v cmd.exe >/dev/null 2>&1; then
  echo "warning: build/check-no-source-binaries.sh is deprecated; delegating to build/check-no-source-binaries.bat" >&2
  exec cmd.exe /c build\\check-no-source-binaries.bat
fi

allow_release_artifacts="${GOAHK_ALLOW_RELEASE_ARTIFACTS:-}"
if [[ "${allow_release_artifacts,,}" == "true" ]]; then
  allow_release_artifacts="1"
fi

tracked_binaries="$(git ls-files '*.exe' '*.dll' '*.so' '*.dylib')"
if [[ -z "${tracked_binaries}" ]]; then
  echo "ok: no blocked tracked binaries detected"
  exit 0
fi

violations=()
while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  if [[ -n "$allow_release_artifacts" && "$file" == dist/releases/* ]]; then
    continue
  fi
  violations+=("$file")
done <<<"${tracked_binaries}"

if [[ ${#violations[@]} -eq 0 ]]; then
  echo "ok: no blocked tracked binaries detected"
  exit 0
fi

echo "error: blocked tracked binaries found. Allowed only in dist/releases/ when GOAHK_ALLOW_RELEASE_ARTIFACTS=1" >&2
printf '%s\n' "${violations[@]}" >&2
exit 1
