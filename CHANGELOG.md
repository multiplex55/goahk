# Changelog

## Unreleased

- Added callback ergonomics on `goahk.Context`: `Err()`, `Sleep(time.Duration) bool`,
  `Logger()`, `Log(...)`, `Binding()`, and `Trigger()`.
- Documented canonical cancellation usage: prefer `ctx.Err()` and `ctx.Sleep(...)`;
  `ctx.Context()` remains available for low-level integrations.
- Updated callback-focused docs/examples to use the supported public callback APIs.
