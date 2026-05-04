@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"

set "OUTDIR=dist\goahk-uia-viewer"
set "OUTEXE=%OUTDIR%\goahk-uia-viewer.exe"
set "MANIFEST=cmd\goahk-uia-viewer\goahk-uia-viewer.manifest"
set "RSRC=cmd\goahk-uia-viewer\goahk-uia-viewer_windows.syso"

pushd "%ROOT%" || exit /b 1

if not exist "%OUTDIR%" mkdir "%OUTDIR%"
if errorlevel 1 goto :fail

if exist "%RSRC%" del /f /q "%RSRC%"
if errorlevel 1 goto :fail

go run github.com/akavel/rsrc@latest -manifest "%MANIFEST%" -o "%RSRC%"
if errorlevel 1 goto :fail

set "LDFLAGS="
if "%GOAHK_UIA_VIEWER_RELEASE%"=="1" set "LDFLAGS=-H=windowsgui"

if defined LDFLAGS (
  go build -trimpath -v -ldflags "%LDFLAGS%" -o "%OUTEXE%" ./cmd/goahk-uia-viewer
) else (
  go build -trimpath -v -o "%OUTEXE%" ./cmd/goahk-uia-viewer
)
if errorlevel 1 goto :fail

echo Built %OUTEXE%
set "EXIT_CODE=0"
goto :cleanup

:fail
set "EXIT_CODE=%ERRORLEVEL%"

:cleanup
if exist "%RSRC%" del /f /q "%RSRC%" >nul 2>nul
popd
exit /b %EXIT_CODE%
