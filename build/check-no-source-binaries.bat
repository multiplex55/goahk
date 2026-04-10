@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "ALLOW_RELEASE_ARTIFACTS=%GOAHK_ALLOW_RELEASE_ARTIFACTS%"
if /I "%ALLOW_RELEASE_ARTIFACTS%"=="true" set "ALLOW_RELEASE_ARTIFACTS=1"

set "HAS_VIOLATIONS="
for /f "usebackq delims=" %%I in (`git ls-files`) do (
  echo %%I | findstr /R /I "\.exe$ \.dll$ \.so$ \.dylib$" >NUL
  if not errorlevel 1 (
    set "IS_APPROVED="
    if defined ALLOW_RELEASE_ARTIFACTS (
      echo %%I | findstr /B /I "dist/releases/ dist\releases\" >NUL
      if not errorlevel 1 set "IS_APPROVED=1"
    )

    if not defined IS_APPROVED (
      if not defined HAS_VIOLATIONS (
        >&2 echo error: blocked tracked binaries found. Allowed only in dist/releases/ when GOAHK_ALLOW_RELEASE_ARTIFACTS=1
        set "HAS_VIOLATIONS=1"
      )
      >&2 echo %%I
    )
  )
)

if defined HAS_VIOLATIONS exit /b 1

echo ok: no blocked tracked binaries detected
exit /b 0
