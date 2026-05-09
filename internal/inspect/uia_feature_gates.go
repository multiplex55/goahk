package inspect

import (
	"log"
	"time"
)

type UIAFeatureGates struct {
	DisablePatterns   bool
	MinimalProperties bool
	DisableAutoExpand bool
	MaxInitialDepth   int
	MaxInitialNodes   int
	BranchTimeout     time.Duration
	TotalLoadTimeout  time.Duration
	TraceCOM          bool
}

func (g UIAFeatureGates) normalized() UIAFeatureGates {
	if g.MaxInitialDepth < 1 {
		g.MaxInitialDepth = 3
	}
	if g.MaxInitialDepth > 8 {
		g.MaxInitialDepth = 8
	}
	if g.MaxInitialNodes < 1 {
		g.MaxInitialNodes = 300
	}
	if g.MaxInitialNodes > 5000 {
		g.MaxInitialNodes = 5000
	}
	return g
}

var activeUIAFeatureGates = (UIAFeatureGates{}).normalized()

func SetUIAFeatureGates(gates UIAFeatureGates) {
	activeUIAFeatureGates = gates.normalized()
	log.Printf("inspect.uia.feature_gates disable_patterns=%t minimal_properties=%t disable_auto_expand=%t max_initial_depth=%d max_initial_nodes=%d branch_timeout=%s total_load_timeout=%s trace_com=%t", activeUIAFeatureGates.DisablePatterns, activeUIAFeatureGates.MinimalProperties, activeUIAFeatureGates.DisableAutoExpand, activeUIAFeatureGates.MaxInitialDepth, activeUIAFeatureGates.MaxInitialNodes, activeUIAFeatureGates.BranchTimeout, activeUIAFeatureGates.TotalLoadTimeout, activeUIAFeatureGates.TraceCOM)
}

func GetUIAFeatureGates() UIAFeatureGates { return activeUIAFeatureGates }
