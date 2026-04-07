@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"
set "OUT_DIR=%ROOT%\dist"
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

if not defined VERSION set "VERSION=v0.1.0"

if not defined COMMIT (
  for /f "usebackq delims=" %%I in (`git -C "%ROOT%" rev-parse --short=7 HEAD 2^>NUL`) do set "COMMIT=%%I"
  if not defined COMMIT set "COMMIT=unknown"
)

if not defined SOURCE_DATE_EPOCH (
  for /f "usebackq delims=" %%I in (`powershell -NoProfile -Command "[DateTimeOffset]::UtcNow.ToUnixTimeSeconds()"`) do set "SOURCE_DATE_EPOCH=%%I"
)

for /f "usebackq delims=" %%I in (`powershell -NoProfile -Command "$ErrorActionPreference = 'Stop'; [DateTimeOffset]::FromUnixTimeSeconds([int64]$env:SOURCE_DATE_EPOCH).UtcDateTime.ToString('yyyy-MM-ddTHH:mm:ssZ')"`) do set "BUILD_DATE=%%I"
if errorlevel 1 (
  >&2 echo error: invalid SOURCE_DATE_EPOCH value "%SOURCE_DATE_EPOCH%"
  exit /b 1
)
if not defined BUILD_DATE (
  >&2 echo error: failed to compute BUILD_DATE
  exit /b 1
)

set "LDFLAGS=-X main.version=%VERSION% -X main.commit=%COMMIT% -X main.buildDate=%BUILD_DATE%"

go build -trimpath -buildvcs=false -ldflags "%LDFLAGS%" -o "%OUT_DIR%\goahk" "%ROOT%\cmd\goahk"
if errorlevel 1 exit /b 1

exit /b 0
