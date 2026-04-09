package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/flow"
	"goahk/internal/runtime"
)

func main() {
	configPath := flag.String("config", "testdata/config/valid_minimal.json", "path to config fixture")
	bindingID := flag.String("binding", "", "binding id to execute")
	flag.Parse()

	p, err := config.LoadProgramFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load program: %v\n", err)
		os.Exit(1)
	}
	registry := actions.NewRegistry()
	compiled, err := runtime.CompileRuntimeBindings(p, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile bindings: %v\n", err)
		os.Exit(1)
	}
	executor := actions.NewExecutor(registry)
	for _, b := range compiled {
		if *bindingID != "" && b.ID != *bindingID {
			continue
		}
		if b.Flow != nil {
			res := executor.ExecuteFlow(actions.ActionContext{Context: context.Background()}, *b.Flow, flow.ConditionEvaluator{})
			fmt.Printf("%s flow success=%v steps=%d\n", b.ID, res.Success, len(res.Steps))
			continue
		}
		res := executor.Execute(actions.ActionContext{Context: context.Background()}, b.Plan)
		fmt.Printf("%s success=%v steps=%d\n", b.ID, res.Success, len(res.Steps))
	}
}
