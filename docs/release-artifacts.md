# Release artifact strategy

Binary release artifacts must not be committed into the source tree.

## Distribution channels

Use one of the following channels for compiled binaries:

1. **GitHub Releases**
   - Publish versioned binaries as release assets.
   - Keep checksums/signatures alongside release assets where applicable.
2. **CI/CD pipeline artifacts**
   - Upload build outputs from CI as ephemeral or retained artifacts.
   - Promote signed artifacts to release assets when cutting a release.

## Repository hygiene policy

- Source control tracks source code, tests, scripts, and documentation only.
- Generated binaries (`.exe`, `.dll`, `.so`, `.dylib`) are blocked by:
  - `.gitignore` patterns.
  - CI guard (`build/check-no-source-binaries.*`).
  - Repohygiene unit test (`internal/repohygiene/tracked_artifacts_test.go`).

