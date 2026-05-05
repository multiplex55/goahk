@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
for %%I in ("%ROOT%") do set "ROOT=%%~fI"

pushd "%ROOT%" || exit /b 1
go test -v ./internal/repohygiene -run TestNoTrackedBinaryArtifacts -count=1
if errorlevel 1 goto :fail

go test -v ./build -run TestViewerDirectoryHasNoFrontendArtifacts -count=1
if errorlevel 1 goto :fail

set "EXIT_CODE=0"
goto :done

:fail
set "EXIT_CODE=%ERRORLEVEL%"

:done
popd

exit /b %EXIT_CODE%
