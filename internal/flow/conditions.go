package flow

import "context"

type Condition struct {
	WindowMatches *WindowCondition
	ElementExists *ElementCondition
}

type WindowCondition struct {
	Matcher string
}

type ElementCondition struct {
	Selector string
}

type WindowEvaluator interface {
	WindowMatches(context.Context, string) (bool, error)
}

type ElementEvaluator interface {
	ElementExists(context.Context, string) (bool, error)
}

type ConditionEvaluator struct {
	Windows  WindowEvaluator
	Elements ElementEvaluator
}

func (e ConditionEvaluator) Evaluate(ctx context.Context, cond Condition) (bool, error) {
	if cond.WindowMatches != nil {
		if e.Windows == nil {
			return false, nil
		}
		return e.Windows.WindowMatches(ctx, cond.WindowMatches.Matcher)
	}
	if cond.ElementExists != nil {
		if e.Elements == nil {
			return false, nil
		}
		return e.Elements.ElementExists(ctx, cond.ElementExists.Selector)
	}
	return false, nil
}
