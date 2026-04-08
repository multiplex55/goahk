# standalone-script

This example is intended to be compiled and run as a standalone executable.

## Behavior

- Press `1` to show a Windows MessageBox (`"You pressed 1"`).
- Press `Escape` to stop the app and exit the process.

The app keeps running until the `Escape` keybind is triggered.

## Build

```powershell
go build -o standalone-script.exe ./examples/standalone-script
```

## Run

```powershell
.\standalone-script.exe
```
