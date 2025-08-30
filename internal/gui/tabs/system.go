package tabs

import (
	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/dialogs" // New import

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type SystemTab struct {
	fyne.CanvasObject
	refreshPendingActions func()
}

func (t *SystemTab) Refresh() {
	t.refreshPendingActions()
}

// CreateSystemTab creates the content for the "System" tab
func CreateSystemTab(
	runCmd func(f func() (string, error), msg string, showOutput bool, refresh func()),
	flakePath string,
	prefs fyne.Preferences,
	w fyne.Window, // Add w here
	refreshPendingActions func(),
) *SystemTab {
	rebuildButton := widget.NewButton("🚀  Commit & Rebuild", func() {
		dialogs.ShowPasswordDialog(w, func(password string) {
			runCmd(func() (string, error) {
				out, err := api.Rebuild(flakePath, password, "", "")
				if err != nil {
					config.AddLogEntry("Error rebuilding system: " + err.Error())
					return out, err
				}
				config.AddLogEntry("System rebuilt successfully!")

				// Add and commit changes after successful rebuild
				path := config.GetInstallPath()
				if err := api.GitAdd(path); err != nil {
					config.AddLogEntry("Error adding changes: " + err.Error())
					// Don't return error, just log it
				}
				if err := api.GitCommit(path, "pilo: rebuild"); err != nil {
					config.AddLogEntry("Error committing changes: " + err.Error())
					// Don't return error, just log it
				}

				refreshPendingActions()
				return out, nil
			}, "🚀  Rebuilding system...", true, nil)
		})
	})

	updateButton := widget.NewButton("🔄  Update", func() {
		runCmd(func() (string, error) {
			out, err := api.Update("")
			if err != nil {
				config.AddLogEntry("Error updating flake inputs: " + err.Error())
				return out, err
			}
			config.AddLogEntry("Flake inputs updated successfully!")
			refreshPendingActions() // Call refresh after update
			return out, nil
		}, "🔄  Updating flake inputs...", true, nil)
	})

	rollbackButton := widget.NewButton("↩️  Rollback System", func() {
		dialogs.ShowPasswordDialog(w, func(password string) {
			runCmd(func() (string, error) {
				out, err := api.Rollback(password)
				if err != nil {
					config.AddLogEntry("Error rolling back system: " + err.Error())
					return out, err
				}
				config.AddLogEntry("System rolled back successfully!")
				refreshPendingActions() // Call refresh after rollback
				return out, nil
			}, "↩️  Rolling back to previous generation...", true, nil)
		})
	})

	upgradeButton := widget.NewButton("⬆️  Upgrade Packages", func() {
		runCmd(func() (string, error) {
			out, err := api.Upgrade()
			if err != nil {
				config.AddLogEntry("Error upgrading packages: " + err.Error())
				return out, err
			}
			config.AddLogEntry("Packages upgraded successfully!")
			refreshPendingActions() // Call refresh after upgrade
			return out, nil
		}, "⬆️  Upgrading packages...", true, nil)
	})

	gcButton := widget.NewButton("🗑️  Run Garbage Collection", func() {
		runCmd(func() (string, error) {
			out, err := api.GC()
			if err != nil {
				config.AddLogEntry("Error running garbage collection: " + err.Error())
				return out, err
			}
			config.AddLogEntry("Garbage collection completed successfully!")
			refreshPendingActions() // Call refresh after garbage collection
			return out, nil
		}, "🗑️  Running garbage collector...", true, nil)
	})

	listButton := widget.NewButton("📜  List Generations", func() {
		runCmd(func() (string, error) {
			out, err := api.ListGenerations()
			if err != nil {
				config.AddLogEntry("Error listing generations: " + err.Error())
				return out, err
			}
			config.AddLogEntry("Generations listed successfully!")
			return out, nil
		}, "📜  Listing generations...", true, nil)
	})

	actions := container.NewVBox(
		container.NewGridWithColumns(3,
			container.NewVBox(
				rebuildButton,
				widget.NewLabelWithStyle("Apply pending configuration changes to the system.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
			container.NewVBox(
				updateButton,
				widget.NewLabelWithStyle("Update all flake inputs to the latest version.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
			container.NewVBox(
				rollbackButton,
				widget.NewLabelWithStyle("Roll back to the previous system generation.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			container.NewVBox(
				upgradeButton,
				widget.NewLabelWithStyle("Upgrade all installed packages to the latest version.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
			container.NewVBox(
				gcButton,
				widget.NewLabelWithStyle("Run the garbage collector to free up disk space.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
			container.NewVBox(
				listButton,
				widget.NewLabelWithStyle("List all available system generations.", fyne.TextAlignCenter, fyne.TextStyle{}),
			),
		),
	)

	pendingActionsBinding := binding.NewStringList()
	pendingActionsList := widget.NewListWithData(
		pendingActionsBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			item, _ := i.(binding.String).Get()
			o.(*widget.Label).SetText(item)
		},
	)

	refreshPendingActions()

	topContent := container.NewVBox(
		widget.NewLabelWithStyle("System Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		actions,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Pending Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	content := container.NewBorder(
		topContent,
		nil,
		nil,
		nil,
		pendingActionsList,
	)

	tab := &SystemTab{
		CanvasObject:          container.NewPadded(content),
		refreshPendingActions: refreshPendingActions,
	}

	go func() {
		dirty, err := api.GitStatus(config.GetInstallPath())
		if err != nil {
			fyne.LogError("Failed to get git status for initial rebuild button importance", err)
			return
		}
		if dirty {
			rebuildButton.Importance = widget.HighImportance
		} else {
			rebuildButton.Importance = widget.MediumImportance
		}
		rebuildButton.Refresh()
	}()

	return tab
}
