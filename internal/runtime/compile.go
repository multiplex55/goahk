package runtime

import (
	"goahk/internal/actions"
	appinternal "goahk/internal/app"
	"goahk/internal/program"
)

type RuntimeBinding = appinternal.RuntimeBinding

// CompileRuntimeBindings compiles a canonical program.Program into runtime bindings.
func CompileRuntimeBindings(p program.Program, registry *actions.Registry) ([]RuntimeBinding, error) {
	return appinternal.CompileRuntimeBindingsFromProgram(p, registry)
}
