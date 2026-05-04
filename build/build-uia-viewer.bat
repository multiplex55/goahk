@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"
set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"
set "UIA_DIR=%ROOT%\cmd\goahk-uia-viewer"
set "MANIFEST=%UIA_DIR%\goahk-uia-viewer.manifest"
set "MANIFEST_SYSO=%UIA_DIR%\zz_windows_manifest.syso"

if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"

pushd "%ROOT%" || exit /b 1
if exist "%MANIFEST%" (
  go run github.com/akavel/rsrc@v0.10.2 -manifest "%MANIFEST%" -o "%MANIFEST_SYSO%"
  if errorlevel 1 (
    set "EXIT_CODE=%ERRORLEVEL%"
    popd
    exit /b %EXIT_CODE%
  )
)
go build -trimpath -v -o "%DIST_DIR%\goahk-uia-viewer.exe" ./cmd/goahk-uia-viewer
set "EXIT_CODE=%ERRORLEVEL%"
if exist "%MANIFEST_SYSO%" del /f /q "%MANIFEST_SYSO%" >nul 2>nul
popd

exit /b %EXIT_CODE%
