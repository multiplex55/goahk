package program

type StepSpec struct {
	Action string
	Params map[string]any
	When   *ConditionSpec
}

type ConditionSpec struct {
	Expression string
}
