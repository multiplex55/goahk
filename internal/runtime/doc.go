// Package runtime is the canonical runtime lifecycle and compile path for goahk v1.
//
// Canonical compile entrypoint:
//   - CompileRuntimeBindings(program.Program, *actions.Registry)
//
// Canonical execution entrypoint:
//   - Bootstrap.Run / Bootstrap.RunProgram
//
// Entry points should route through this package directly, with config-to-program
// adaptation performed in internal/config.
package runtime
