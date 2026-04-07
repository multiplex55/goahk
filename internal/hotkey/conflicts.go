package hotkey

import (
	"fmt"
	"sort"
	"strings"
)

type ConflictDetail struct {
	Chord      string
	BindingIDs []string
	Suggestion string
}

type ConflictReport struct {
	Conflicts []ConflictDetail
}

func (r ConflictReport) Error() string {
	if len(r.Conflicts) == 0 {
		return ""
	}
	first := r.Conflicts[0]
	return fmt.Sprintf("hotkey conflict: %q and %q both map to %s", first.BindingIDs[0], first.BindingIDs[1], first.Chord)
}

func (r ConflictReport) UserMessage() string {
	if len(r.Conflicts) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Hotkey conflicts detected:\n")
	for _, c := range r.Conflicts {
		b.WriteString(fmt.Sprintf("- %s used by [%s]. %s\n", c.Chord, strings.Join(c.BindingIDs, ", "), c.Suggestion))
	}
	return strings.TrimSpace(b.String())
}

func DetectConflictReport(bindings []Binding) *ConflictReport {
	grouped := map[string][]string{}
	for _, b := range bindings {
		norm := b.Chord.String()
		grouped[norm] = append(grouped[norm], b.ID)
	}
	chords := make([]string, 0, len(grouped))
	for chord, ids := range grouped {
		if len(ids) > 1 {
			chords = append(chords, chord)
		}
	}
	sort.Strings(chords)
	if len(chords) == 0 {
		return nil
	}
	report := &ConflictReport{Conflicts: make([]ConflictDetail, 0, len(chords))}
	for _, chord := range chords {
		ids := append([]string(nil), grouped[chord]...)
		sort.Strings(ids)
		report.Conflicts = append(report.Conflicts, ConflictDetail{
			Chord:      chord,
			BindingIDs: ids,
			Suggestion: fmt.Sprintf("choose unique chords (for example keep %q and remap the others)", ids[0]),
		})
	}
	return report
}

func DetectConflicts(bindings []Binding) error {
	report := DetectConflictReport(bindings)
	if report == nil {
		return nil
	}
	return report
}
