# Build and test guide

This page is the quick navigation point for building and validating `goahk`.

## Build

- Packaging/build script and version metadata live under [`build/`](../build).
- Primary build command:

```bash
./build/build.sh
```

For packaging details and emitted artifacts, see [`build/README.md`](../build/README.md).

## Test

Common commands:

```bash
go test ./...
go test -race ./...
```

For full strategy (including Windows integration tags), see [`docs/testing.md`](./testing.md).

## See also

- Runtime usage and configuration examples: [`docs/USAGE.md`](./USAGE.md)
- Project overview: [`README.md`](../README.md)
