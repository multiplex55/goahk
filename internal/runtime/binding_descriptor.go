package runtime

import (
	"fmt"
	"strings"

	"goahk/internal/actions"
)

func buildExecutableBindings(compiled []RuntimeBinding) map[string]actions.ExecutableBinding {
	descriptors := make(map[string]actions.ExecutableBinding, len(compiled))
	for _, binding := range compiled {
		descriptor := actions.ExecutableBinding{ID: binding.ID, Kind: actions.BindingKindPlan, Plan: binding.Plan}
		switch {
		case binding.Flow != nil:
			descriptor.Kind = actions.BindingKindFlow
			descriptor.Flow = binding.Flow
		case isCallbackPlan(binding.Plan):
			descriptor.Kind = actions.BindingKindCallback
			descriptor.Policy.CallbackRef = strings.TrimSpace(binding.Plan[0].Params["callback_ref"])
		}
		descriptors[binding.ID] = descriptor
	}
	return descriptors
}

func validateExecutableBinding(binding actions.ExecutableBinding) error {
	switch binding.Kind {
	case actions.BindingKindPlan:
		return nil
	case actions.BindingKindFlow:
		if binding.Flow == nil {
			return fmt.Errorf("binding descriptor kind \"flow\" requires compiled flow payload")
		}
		return nil
	case actions.BindingKindCallback:
		return nil
	default:
		return fmt.Errorf("invalid binding descriptor kind %q", binding.Kind)
	}
}

func isCallbackPlan(plan actions.Plan) bool {
	if len(plan) != 1 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(plan[0].Name), actions.CallbackActionName)
}
