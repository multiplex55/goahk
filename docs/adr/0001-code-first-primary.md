# ADR 0001: Code-first API is the primary authoring model

- **Status:** Accepted
- **Date:** 2026-04-08

## Context

`goahk` supports two entry points today:

1. Go-authored scripts using the `goahk` package.
2. JSON configuration loaded by the compatibility runner (`cmd/goahk`).

We need a stable statement of intent so docs, examples, tests, and release notes present one clear default while preserving migration safety for existing JSON users.

## Decision

- **Primary mode:** Go-authored script API (`goahk` package) is the default and recommended path for all new automation.
- **Compatibility mode:** JSON config runner (`cmd/goahk`) remains supported as an adapter over the same runtime model for existing deployments.

## Non-goals

- Do not remove JSON config support in this ADR.
- Do not introduce a new DSL or third authoring surface.
- Do not guarantee feature parity for workflows that require custom Go callbacks (`goahk.Func`) because JSON does not encode arbitrary Go code.

## Migration boundaries

- Existing JSON configurations are expected to continue to run through `cmd/goahk`.
- New examples, quick-starts, and feature narratives should lead with the code-first API.
- Compatibility guarantees are defined only for features represented in both surfaces and covered by parity tests.

## Consequences

- Architecture and usage docs must place code-first first, with JSON presented as compatibility/adapter mode.
- Release notes must separately call out:
  - code-first API changes
  - config compatibility updates
