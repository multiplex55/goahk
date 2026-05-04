@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"
set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"

if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"

pushd "%ROOT%" || exit /b 1
go build -trimpath -v -o "%DIST_DIR%\goahk-uia-viewer.exe" ./cmd/goahk-uia-viewer
set "EXIT_CODE=%ERRORLEVEL%"
popd

exit /b %EXIT_CODE%
