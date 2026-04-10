package goahk

import "goahk/internal/uia"

type Selector struct {
	AutomationID string
	Name         string
	ControlType  string
	Ancestors    []Selector
}

func SelectByAutomationID(id string) Selector  { return Selector{AutomationID: id} }
func SelectByName(name string) Selector        { return Selector{Name: name} }
func SelectByControlType(kind string) Selector { return Selector{ControlType: kind} }

func (s Selector) WithAutomationID(id string) Selector {
	s.AutomationID = id
	return s
}

func (s Selector) WithName(name string) Selector {
	s.Name = name
	return s
}

func (s Selector) WithControlType(kind string) Selector {
	s.ControlType = kind
	return s
}

func (s Selector) WithAncestors(ancestors ...Selector) Selector {
	s.Ancestors = append([]Selector(nil), ancestors...)
	return s
}

func (s Selector) toInternal() uia.Selector {
	out := uia.Selector{AutomationID: s.AutomationID, Name: s.Name, ControlType: s.ControlType}
	if len(s.Ancestors) == 0 {
		return out
	}
	out.Ancestors = make([]uia.Selector, 0, len(s.Ancestors))
	for _, anc := range s.Ancestors {
		out.Ancestors = append(out.Ancestors, anc.toInternal())
	}
	return out
}
