//go:build windows
// +build windows

package folders

import (
	"context"
	"fmt"
	"os/exec"
)

type windowsService struct{}

func newPlatformService() Service {
	return windowsService{}
}

func (windowsService) ListOpenFolders(ctx context.Context) ([]FolderInfo, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", listFoldersScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("folders: enumerate explorer windows: %w", err)
	}
	parsed, err := parsePowerShellFolderResults(out)
	if err != nil {
		return nil, err
	}
	return normalizeAndDedupe(parsed, false), nil
}

const listFoldersScript = `$ErrorActionPreference = 'Stop'
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public static class Win32Folders {
    [DllImport("user32.dll", SetLastError=true)]
    public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
}
"@

$windows = New-Object -ComObject Shell.Application
$results = @()
foreach ($window in $windows.Windows()) {
    try {
        $hwnd = [int64]$window.HWND
        if ($hwnd -eq 0) { continue }

        $folder = $window.Document.Folder
        if ($null -eq $folder -or $null -eq $folder.Self) { continue }

        $path = [string]$folder.Self.Path
        if ([string]::IsNullOrWhiteSpace($path)) {
            continue
        }

        $pid = [uint32]0
        [void][Win32Folders]::GetWindowThreadProcessId([IntPtr]$hwnd, [ref]$pid)

        $results += [pscustomobject]@{
            path  = $path
            title = [string]$window.LocationName
            pid   = $pid
            hwnd  = ('0x{0:X}' -f $hwnd)
        }
    }
    catch {
        $results += [pscustomobject]@{
            path       = ''
            title      = ''
            pid        = [uint32]0
            hwnd       = ''
            diagnostic = $_.Exception.Message
        }
    }
}

$results | ConvertTo-Json -Depth 4 -Compress`
