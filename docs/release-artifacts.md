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


## PR and workflow exceptions

To keep automation predictable, binary artifacts are only allowed when all of the following are true:

1. The PR is dedicated to release artifacts (branch prefix: `release-artifacts/`).
2. The PR includes the `release-artifacts` label.
3. Any committed binaries are under `dist/releases/`.

CI enforcement details:

- `build/check-no-source-binaries.*` blocks tracked binaries outside `dist/releases/`.
- The exception gate is environment variable `GOAHK_ALLOW_RELEASE_ARTIFACTS=1`.
- GitHub Actions sets that variable only for PRs carrying the `release-artifacts` label.

This means normal feature/fix PRs fail fast if generated binaries are committed, while release PRs remain explicit and reviewable.
