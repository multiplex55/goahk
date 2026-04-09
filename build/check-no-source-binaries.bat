@echo off
setlocal EnableExtensions

set "HAS_TRACKED_BINARIES="
for /f "usebackq delims=" %%I in (`git ls-files ^| findstr /R /I "\.\(exe\|dll\|so\|dylib\)$"`) do (
  if not defined HAS_TRACKED_BINARIES (
    >&2 echo error: tracked binary artifacts are not allowed (.exe/.dll/.so/.dylib):
    set "HAS_TRACKED_BINARIES=1"
  )
  >&2 echo %%I
)

if defined HAS_TRACKED_BINARIES exit /b 1

echo ok: no tracked binary artifacts (.exe/.dll/.so/.dylib)
exit /b 0
