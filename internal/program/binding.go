package program

type BindingSpec struct {
	ID     string
	Hotkey string
	Steps  []StepSpec
	Flow   string
}
