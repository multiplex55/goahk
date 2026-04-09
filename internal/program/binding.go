package program

type BindingSpec struct {
	ID                string
	Hotkey            string
	Steps             []StepSpec
	Flow              string
	ConcurrencyPolicy ConcurrencyPolicy
}

type ConcurrencyPolicy string

const (
	ConcurrencyPolicySerial   ConcurrencyPolicy = "serial"
	ConcurrencyPolicyReplace  ConcurrencyPolicy = "replace"
	ConcurrencyPolicyParallel ConcurrencyPolicy = "parallel"
	ConcurrencyPolicyQueueOne ConcurrencyPolicy = "queue-one"
	ConcurrencyPolicyDrop     ConcurrencyPolicy = "drop"
)

func DefaultConcurrencyPolicy() ConcurrencyPolicy {
	return ConcurrencyPolicySerial
}
