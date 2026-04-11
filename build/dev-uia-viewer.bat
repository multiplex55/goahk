@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"
set "APP_DIR=%ROOT%\cmd\goahk-uia-viewer"
set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"

if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

pushd "%APP_DIR%" || exit /b 1
wails dev
set "EXIT_CODE=%ERRORLEVEL%"
popd

exit /b %EXIT_CODE%
