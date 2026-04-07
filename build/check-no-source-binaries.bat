@echo off
setlocal EnableExtensions

set "HAS_TRACKED_EXE="
for /f "usebackq delims=" %%I in (`git ls-files "*.exe"`) do (
  if not defined HAS_TRACKED_EXE (
    >&2 echo error: tracked .exe artifacts are not allowed:
    set "HAS_TRACKED_EXE=1"
  )
  >&2 echo %%I
)

if defined HAS_TRACKED_EXE exit /b 1

echo ok: no tracked .exe artifacts
exit /b 0
