@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"
set "APP_DIR=%ROOT%\cmd\goahk-uia-viewer"
set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"

if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"

pushd "%APP_DIR%" || exit /b 1
wails build -clean -o goahk-uia-viewer
if errorlevel 1 (
  set "EXIT_CODE=%ERRORLEVEL%"
  popd
  exit /b %EXIT_CODE%
)

if exist "build\bin" (
  robocopy "build\bin" "%DIST_DIR%" /E >NUL
  if errorlevel 8 (
    set "EXIT_CODE=%ERRORLEVEL%"
    popd
    exit /b %EXIT_CODE%
  )
)

popd
exit /b 0
