package inspect

import "log"

type UIAFeatureGates struct {
	DisablePatterns   bool
	MinimalProperties bool
	DisableAutoExpand bool
	MaxInitialDepth   int
	TraceCOM          bool
}

func (g UIAFeatureGates) normalized() UIAFeatureGates {
	if g.MaxInitialDepth < 1 {
		g.MaxInitialDepth = 3
	}
	if g.MaxInitialDepth > 8 {
		g.MaxInitialDepth = 8
	}
	return g
}

var activeUIAFeatureGates = (UIAFeatureGates{}).normalized()

func SetUIAFeatureGates(gates UIAFeatureGates) {
	activeUIAFeatureGates = gates.normalized()
	log.Printf("inspect.uia.feature_gates disable_patterns=%t minimal_properties=%t disable_auto_expand=%t max_initial_depth=%d trace_com=%t", activeUIAFeatureGates.DisablePatterns, activeUIAFeatureGates.MinimalProperties, activeUIAFeatureGates.DisableAutoExpand, activeUIAFeatureGates.MaxInitialDepth, activeUIAFeatureGates.TraceCOM)
}

func GetUIAFeatureGates() UIAFeatureGates { return activeUIAFeatureGates }
