Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

Describe 'build.bat' {
    BeforeAll {
        $RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
        $BuildBatSource = Join-Path $RepoRoot 'build\build.bat'
    }

    function New-BuildFixture {
        $fixtureRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("goahk-build-fixture-{0}" -f ([guid]::NewGuid().ToString('N')))
        New-Item -ItemType Directory -Path $fixtureRoot | Out-Null
        New-Item -ItemType Directory -Path (Join-Path $fixtureRoot 'build') | Out-Null
        New-Item -ItemType Directory -Path (Join-Path $fixtureRoot 'cmd\goahk') -Force | Out-Null

        Copy-Item -Path $BuildBatSource -Destination (Join-Path $fixtureRoot 'build\build.bat')

        @'
module fixture

go 1.22
'@ | Set-Content -Path (Join-Path $fixtureRoot 'go.mod') -NoNewline

        @'
package main

import "log"

var (
    version = "dev"
    commit = "unknown"
    buildDate = "unknown"
)

func main() {
    log.Printf("version=%s commit=%s buildDate=%s", version, commit, buildDate)
}
'@ | Set-Content -Path (Join-Path $fixtureRoot 'cmd\goahk\main.go') -NoNewline

        Push-Location $fixtureRoot
        try {
            git init | Out-Null
            git config user.email 'fixture@example.com'
            git config user.name 'Fixture User'
            git add .
            git commit -m 'fixture init' | Out-Null
        }
        finally {
            Pop-Location
        }

        return $fixtureRoot
    }

    function Invoke-CmdScript {
        param(
            [Parameter(Mandatory)]
            [string]$WorkingDirectory,
            [Parameter(Mandatory)]
            [string]$Command,
            [hashtable]$EnvVars
        )

        $envCopy = @{}
        if ($EnvVars) {
            foreach ($entry in $EnvVars.GetEnumerator()) {
                $name = [string]$entry.Key
                $envCopy[$name] = [Environment]::GetEnvironmentVariable($name)
                [Environment]::SetEnvironmentVariable($name, [string]$entry.Value)
            }
        }

        Push-Location $WorkingDirectory
        try {
            $output = (& cmd.exe /c $Command 2>&1) | Out-String
            return [PSCustomObject]@{
                ExitCode = $LASTEXITCODE
                Output   = $output
            }
        }
        finally {
            Pop-Location
            if ($EnvVars) {
                foreach ($entry in $EnvVars.GetEnumerator()) {
                    $name = [string]$entry.Key
                    [Environment]::SetEnvironmentVariable($name, $envCopy[$name])
                }
            }
        }
    }

    It 'uses default VERSION/COMMIT and generates UTC ISO8601 buildDate when env vars are unset' {
        $fixture = New-BuildFixture
        try {
            $build = Invoke-CmdScript -WorkingDirectory $fixture -Command 'build\build.bat'
            $build.ExitCode | Should -Be 0

            $exe = Join-Path $fixture 'dist\goahk.exe'
            $exeOutput = (& $exe 2>&1) | Out-String

            $exeOutput | Should -Match 'version=v0\.1\.0'
            $exeOutput | Should -Match 'commit=[0-9a-f]{7}'
            $exeOutput | Should -Match 'buildDate=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z'
        }
        finally {
            Remove-Item -Path $fixture -Recurse -Force
        }
    }

    It 'honors explicit VERSION/COMMIT/SOURCE_DATE_EPOCH overrides' {
        $fixture = New-BuildFixture
        try {
            $build = Invoke-CmdScript -WorkingDirectory $fixture -Command 'build\build.bat' -EnvVars @{
                VERSION = 'v9.9.9'
                COMMIT = 'abc1234'
                SOURCE_DATE_EPOCH = '1700000000'
            }
            $build.ExitCode | Should -Be 0

            $exe = Join-Path $fixture 'dist\goahk.exe'
            $exeOutput = (& $exe 2>&1) | Out-String

            $exeOutput | Should -Match 'version=v9\.9\.9'
            $exeOutput | Should -Match 'commit=abc1234'
            $exeOutput | Should -Match 'buildDate=2023-11-14T22:13:20Z'
        }
        finally {
            Remove-Item -Path $fixture -Recurse -Force
        }
    }

    It 'produces deterministic UTC ISO8601 buildDate when SOURCE_DATE_EPOCH is fixed' {
        $fixture = New-BuildFixture
        try {
            $first = Invoke-CmdScript -WorkingDirectory $fixture -Command 'build\build.bat' -EnvVars @{ SOURCE_DATE_EPOCH = '946684800' }
            $first.ExitCode | Should -Be 0
            $firstOutput = (& (Join-Path $fixture 'dist\goahk.exe') 2>&1) | Out-String

            Remove-Item -Path (Join-Path $fixture 'dist') -Recurse -Force

            $second = Invoke-CmdScript -WorkingDirectory $fixture -Command 'build\build.bat' -EnvVars @{ SOURCE_DATE_EPOCH = '946684800' }
            $second.ExitCode | Should -Be 0
            $secondOutput = (& (Join-Path $fixture 'dist\goahk.exe') 2>&1) | Out-String

            $firstDate = [regex]::Match($firstOutput, 'buildDate=(?<d>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)').Groups['d'].Value
            $secondDate = [regex]::Match($secondOutput, 'buildDate=(?<d>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)').Groups['d'].Value

            $firstDate | Should -Be '2000-01-01T00:00:00Z'
            $secondDate | Should -Be $firstDate
        }
        finally {
            Remove-Item -Path $fixture -Recurse -Force
        }
    }
}

Describe 'check-no-source-binaries.bat' {
    BeforeAll {
        $RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
        $CheckBatSource = Join-Path $RepoRoot 'build\check-no-source-binaries.bat'
    }

    function New-CheckFixture {
        param(
            [string[]]$TrackedFiles = @()
        )

        $fixtureRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("goahk-check-fixture-{0}" -f ([guid]::NewGuid().ToString('N')))
        New-Item -ItemType Directory -Path $fixtureRoot | Out-Null
        New-Item -ItemType Directory -Path (Join-Path $fixtureRoot 'build') | Out-Null
        Copy-Item -Path $CheckBatSource -Destination (Join-Path $fixtureRoot 'build\check-no-source-binaries.bat')

        Set-Content -Path (Join-Path $fixtureRoot '.gitignore') -Value '' -NoNewline

        foreach ($relativeFile in $TrackedFiles) {
            $target = Join-Path $fixtureRoot $relativeFile
            New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force | Out-Null
            Set-Content -Path $target -Value 'fixture binary' -NoNewline
        }

        Push-Location $fixtureRoot
        try {
            git init | Out-Null
            git config user.email 'fixture@example.com'
            git config user.name 'Fixture User'
            git add .
            git commit -m 'fixture init' | Out-Null
        }
        finally {
            Pop-Location
        }

        return $fixtureRoot
    }

    It 'returns 0 when no tracked blocked binaries exist' {
        $fixture = New-CheckFixture
        try {
            Push-Location $fixture
            try {
                $result = (& cmd.exe /c 'build\check-no-source-binaries.bat' 2>&1) | Out-String
                $exitCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }

            $exitCode | Should -Be 0
            $result | Should -Match 'ok: no blocked tracked binaries detected'
        }
        finally {
            Remove-Item -Path $fixture -Recurse -Force
        }
    }

    It 'returns non-zero when blocked binaries exist in default mode' {
        $fixture = New-CheckFixture -TrackedFiles @('dist\releases\goahk.exe', 'bin\bad.dll')
        try {
            Push-Location $fixture
            try {
                $result = (& cmd.exe /c 'build\check-no-source-binaries.bat' 2>&1) | Out-String
                $exitCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }

            $exitCode | Should -Not -Be 0
            $result | Should -Match 'error: blocked tracked binaries found'
            $result | Should -Match 'dist/releases/goahk\.exe|dist\\releases\\goahk\.exe'
            $result | Should -Match 'bin/bad\.dll|bin\\bad\.dll'
        }
        finally {
            Remove-Item -Path $fixture -Recurse -Force
        }
    }

    It 'allows dist/releases binaries when release mode is enabled' {
        $fixture = New-CheckFixture -TrackedFiles @('dist\releases\goahk.exe', 'dist\releases\plugin.dll')
        $previous = [Environment]::GetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS')
        try {
            [Environment]::SetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS', '1')
            Push-Location $fixture
            try {
                $result = (& cmd.exe /c 'build\check-no-source-binaries.bat' 2>&1) | Out-String
                $exitCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }

            $exitCode | Should -Be 0
            $result | Should -Match 'ok: no blocked tracked binaries detected'
        }
        finally {
            [Environment]::SetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS', $previous)
            Remove-Item -Path $fixture -Recurse -Force
        }
    }

    It 'still blocks binaries outside dist/releases in release mode' {
        $fixture = New-CheckFixture -TrackedFiles @('dist\releases\goahk.exe', 'bin\bad.so')
        $previous = [Environment]::GetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS')
        try {
            [Environment]::SetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS', 'true')
            Push-Location $fixture
            try {
                $result = (& cmd.exe /c 'build\check-no-source-binaries.bat' 2>&1) | Out-String
                $exitCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }

            $exitCode | Should -Not -Be 0
            $result | Should -Match 'bin/bad\.so|bin\\bad\.so'
        }
        finally {
            [Environment]::SetEnvironmentVariable('GOAHK_ALLOW_RELEASE_ARTIFACTS', $previous)
            Remove-Item -Path $fixture -Recurse -Force
        }
    }
}
