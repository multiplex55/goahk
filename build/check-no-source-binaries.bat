@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"

pushd "%ROOT%" || exit /b 1
go test -v ./internal/repohygiene -run TestNoTrackedBinaryArtifacts -count=1
set "EXIT_CODE=%ERRORLEVEL%"
popd

exit /b %EXIT_CODE%
