# Startup on login

`goahk` can be installed for per-user startup by creating a `Run` entry named `goahk`.

## Install behavior

- Installs a single per-user startup entry (`HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run\\goahk`).
- Command points to the current `goahk.exe` path and includes `-config`.
- Reinstall overwrites the existing `goahk` startup entry.

## Uninstall behavior

- Removes the `goahk` startup entry.
- Does not remove config files or logs.
- Safe to run multiple times; missing entry is treated as already uninstalled.
