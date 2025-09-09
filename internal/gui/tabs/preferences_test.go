package tabs

import (
	"pilo/internal/gui/components"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestPreferencesTab_Refresh(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("Test")
	defer w.Close()

	tab := &PreferencesTab{
		installPathEntry:    components.NewSafeEntry(),
		registryNameEntry:   components.NewSafeEntry(),
		remoteUrlEntry:      components.NewSafeEntry(),
		remoteBranchEntry:   components.NewSafeEntry(),
		pushOnCommitCheck:   widget.NewCheck("", nil),
		nixpkgsEntry:        components.NewSafeEntry(),
		homeManagerEntry:    components.NewSafeEntry(),
		nixInstallCmdEntry:  components.NewSafeEntry(),
		logRetentionEntry:   components.NewSafeEntry(),
		customTerminalEntry: components.NewSafeEntry(),
		systemActionsCheck:  widget.NewCheck("", nil),
		appActionsCheck:     widget.NewCheck("", nil),
		pkgActionsCheck:     widget.NewCheck("", nil),
		CanvasObject:        container.NewVBox(),
	}

	// This will panic if the threading is wrong
	tab.Refresh()
}
