package program

import (
	"fmt"
	"sort"
	"strings"

	"goahk/internal/hotkey"
)

const (
	ErrCodeDuplicateBindingID  = "binding.id.duplicate"
	ErrCodeInvalidHotkey       = "binding.hotkey.invalid"
	ErrCodeConflictingHotkeys  = "binding.hotkey.conflict"
	ErrCodeStepsRequired       = "binding.steps.required"
	ErrCodeStepActionRequired  = "binding.step.action.required"
	ErrCodeUnknownFlow         = "binding.flow.unknown"
	ErrCodeUnknownPolicy       = "binding.policy.unknown"
	ErrCodeControlMixedSteps   = "binding.control.mixed_steps"
	ErrCodeControlFlowConflict = "binding.control.flow_conflict"
	ErrCodeControlParams       = "binding.control.params.forbidden"
)

const (
	controlActionStop     = "runtime.control_stop"
	controlActionHardStop = "runtime.control_hard_stop"
)

type ValidationIssue struct {
	Code    string
	Path    string
	Message string
}

type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, fmt.Sprintf("%s (%s): %s", issue.Code, issue.Path, issue.Message))
	}
	return "invalid program: " + strings.Join(parts, "; ")
}

func (e *ValidationError) HasCode(code string) bool {
	if e == nil {
		return false
	}
	for _, issue := range e.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func Normalize(p Program) Program {
	out := Program{Bindings: make([]BindingSpec, 0, len(p.Bindings)), Options: normalizeOptions(p.Options)}
	for _, b := range p.Bindings {
		out.Bindings = append(out.Bindings, normalizeBinding(b))
	}
	sort.SliceStable(out.Bindings, func(i, j int) bool {
		left := bindingSortKey(out.Bindings[i])
		right := bindingSortKey(out.Bindings[j])
		if left == right {
			return out.Bindings[i].Hotkey < out.Bindings[j].Hotkey
		}
		return left < right
	})
	return out
}

func Validate(p Program) error {
	p = Normalize(p)
	issues := make([]ValidationIssue, 0)
	seenIDs := map[string]struct{}{}
	parsed := make([]hotkey.Binding, 0, len(p.Bindings))

	for i, b := range p.Bindings {
		path := fmt.Sprintf("bindings[%d]", i)
		idKey := bindingSortKey(b)
		if _, exists := seenIDs[idKey]; exists {
			issues = append(issues, ValidationIssue{Code: ErrCodeDuplicateBindingID, Path: path + ".id", Message: fmt.Sprintf("duplicate binding id %q", b.ID)})
			continue
		}
		seenIDs[idKey] = struct{}{}

		binding, err := hotkey.ParseBinding(b.ID, b.Hotkey)
		if err != nil {
			issues = append(issues, ValidationIssue{Code: ErrCodeInvalidHotkey, Path: path + ".hotkey", Message: err.Error()})
			continue
		}
		parsed = append(parsed, binding)

		if strings.TrimSpace(b.Flow) == "" && len(b.Steps) == 0 {
			issues = append(issues, ValidationIssue{Code: ErrCodeStepsRequired, Path: path + ".steps", Message: "steps are required when flow is not set"})
		}
		for j, step := range b.Steps {
			if strings.TrimSpace(step.Action) == "" {
				issues = append(issues, ValidationIssue{Code: ErrCodeStepActionRequired, Path: fmt.Sprintf("%s.steps[%d].action", path, j), Message: "action name is required"})
			}
		}
		if cmdStepIdx, ok := findControlStepIndex(b.Steps); ok {
			if strings.TrimSpace(b.Flow) != "" {
				issues = append(issues, ValidationIssue{
					Code:    ErrCodeControlFlowConflict,
					Path:    path + ".flow",
					Message: "control bindings cannot reference flow",
				})
			}
			if len(b.Steps) != 1 {
				issues = append(issues, ValidationIssue{
					Code:    ErrCodeControlMixedSteps,
					Path:    path + ".steps",
					Message: "control bindings must contain exactly one control action step",
				})
			}
			if len(b.Steps[cmdStepIdx].Params) > 0 {
				issues = append(issues, ValidationIssue{
					Code:    ErrCodeControlParams,
					Path:    fmt.Sprintf("%s.steps[%d].params", path, cmdStepIdx),
					Message: "control action does not accept params",
				})
			}
		}
		if !isAllowedConcurrencyPolicy(b.ConcurrencyPolicy) {
			issues = append(issues, ValidationIssue{Code: ErrCodeUnknownPolicy, Path: path + ".concurrencyPolicy", Message: fmt.Sprintf("unsupported policy %q", b.ConcurrencyPolicy)})
		}
	}
	if err := hotkey.DetectConflicts(parsed); err != nil {
		issues = append(issues, ValidationIssue{Code: ErrCodeConflictingHotkeys, Path: "bindings", Message: err.Error()})
	}

	flowIDs := map[string]struct{}{}
	for _, f := range p.Options.Flows {
		flowIDs[bindingSortKey(BindingSpec{ID: f.ID})] = struct{}{}
	}
	for i, b := range p.Bindings {
		if ref := bindingSortKey(BindingSpec{ID: b.Flow}); ref != "" {
			if _, ok := flowIDs[ref]; !ok {
				issues = append(issues, ValidationIssue{Code: ErrCodeUnknownFlow, Path: fmt.Sprintf("bindings[%d].flow", i), Message: fmt.Sprintf("binding %q references unknown flow %q", b.ID, b.Flow)})
			}
		}
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

func normalizeOptions(in Options) Options {
	out := Options{
		Flows:                        make([]FlowSpec, 0, len(in.Flows)),
		UIASelectors:                 make(map[string]UIASelectorSpec, len(in.UIASelectors)),
		EnableImplicitEscapeControls: in.EnableImplicitEscapeControls,
	}
	for _, f := range in.Flows {
		out.Flows = append(out.Flows, normalizeFlow(f))
	}
	sort.SliceStable(out.Flows, func(i, j int) bool {
		return bindingSortKey(BindingSpec{ID: out.Flows[i].ID}) < bindingSortKey(BindingSpec{ID: out.Flows[j].ID})
	})
	for k, v := range in.UIASelectors {
		out.UIASelectors[k] = v
	}
	return out
}

func normalizeFlow(in FlowSpec) FlowSpec {
	out := in
	out.ID = strings.TrimSpace(in.ID)
	if len(in.Steps) > 0 {
		out.Steps = make([]FlowStepSpec, len(in.Steps))
		copy(out.Steps, in.Steps)
	}
	return out
}

func normalizeBinding(in BindingSpec) BindingSpec {
	out := BindingSpec{
		ID:                strings.TrimSpace(in.ID),
		Hotkey:            strings.TrimSpace(in.Hotkey),
		Flow:              strings.TrimSpace(in.Flow),
		ConcurrencyPolicy: normalizeConcurrencyPolicy(in.ConcurrencyPolicy),
	}
	if parsed, err := hotkey.ParseBinding(out.ID, out.Hotkey); err == nil {
		out.Hotkey = parsed.Chord.String()
	}
	if len(in.Steps) > 0 {
		out.Steps = make([]StepSpec, 0, len(in.Steps))
		for _, step := range in.Steps {
			out.Steps = append(out.Steps, normalizeStep(step))
		}
	}
	return out
}

func normalizeStep(in StepSpec) StepSpec {
	out := StepSpec{Action: strings.TrimSpace(in.Action)}
	if len(in.Params) == 0 {
		out.Params = map[string]any{}
		return out
	}
	out.Params = make(map[string]any, len(in.Params))
	for k, v := range in.Params {
		out.Params[strings.TrimSpace(k)] = v
	}
	return out
}

func bindingSortKey(b BindingSpec) string {
	return strings.ToLower(strings.TrimSpace(b.ID))
}

func normalizeConcurrencyPolicy(in ConcurrencyPolicy) ConcurrencyPolicy {
	p := ConcurrencyPolicy(strings.ToLower(strings.TrimSpace(string(in))))
	if p == "" {
		return DefaultConcurrencyPolicy()
	}
	return p
}

func isAllowedConcurrencyPolicy(p ConcurrencyPolicy) bool {
	switch p {
	case ConcurrencyPolicySerial, ConcurrencyPolicyReplace, ConcurrencyPolicyParallel, ConcurrencyPolicyQueueOne, ConcurrencyPolicyDrop:
		return true
	default:
		return false
	}
}

func findControlStepIndex(steps []StepSpec) (int, bool) {
	for idx, step := range steps {
		switch strings.ToLower(strings.TrimSpace(step.Action)) {
		case controlActionStop, controlActionHardStop:
			return idx, true
		}
	}
	return -1, false
}
