package hotkey

import "time"

type ListenerEvent struct {
	RegistrationID int
	TriggeredAt    time.Time
}

type TriggerEvent struct {
	BindingID   string
	Chord       Chord
	TriggeredAt time.Time
}
