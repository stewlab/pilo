package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SafeEntry is a wrapper around widget.Entry that prevents a crash on right-click
// when the entry is not yet part of a canvas.
type SafeEntry struct {
	widget.Entry
	canvas fyne.Canvas
}

// TappedSecondary is a workaround for a bug in Fyne that causes a crash when
// an entry is right-clicked before it has a canvas.
func (e *SafeEntry) TappedSecondary(pe *fyne.PointEvent) {
	if e.canvas == nil {
		return // Canvas not set, do nothing to prevent crash
	}
	e.Entry.TappedSecondary(pe)
}

// NewSafeEntry creates a new SafeEntry widget.
func NewSafeEntry() *SafeEntry {
	entry := &SafeEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}
