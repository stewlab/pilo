package tabs

import (
	"pilo/internal/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"pilo/internal/dialogs"
)

type DevshellTab struct {
	fyne.CanvasObject
	list      *widget.List
	devshells []api.Devshell
}

func (t *DevshellTab) Refresh() {
	t.devshells, _ = api.ListDevshells()
	t.list.Refresh()
}

func CreateDevshellTab(runCmd func(func() error, string, bool, func()), flakePath string, w fyne.Window, refreshPendingActions func()) *DevshellTab {
	tab := &DevshellTab{}

	var err error
	tab.devshells, err = api.ListDevshells()
	if err != nil {
		// handle error
	}

	tab.list = widget.NewList(
		func() int {
			return len(tab.devshells)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("template"),
				layout.NewSpacer(),
				widget.NewButton("...", nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			shellName := tab.devshells[i].Name
			hbox := o.(*fyne.Container)
			label := hbox.Objects[0].(*widget.Label)
			label.SetText(shellName)

			button := hbox.Objects[2].(*widget.Button)
			button.OnTapped = func() {
				menu := fyne.NewMenu("",
					fyne.NewMenuItem("▶️  Enter", func() {
						runCmd(func() error {
							return api.EnterDevshell(shellName, flakePath)
						}, "▶️  Entering devshell...", false, nil)
					}),
					fyne.NewMenuItem("✏️  Edit", func() {
						content, err := api.GetDevshellContent(shellName)
						if err != nil {
							dialogs.ShowErrorDialog(err, w)
							return
						}
						fileNameEntry := widget.NewEntry()
						fileNameEntry.SetText(shellName)

						contentEntry := widget.NewMultiLineEntry()
						contentEntry.SetText(content)
						contentScroll := container.NewScroll(contentEntry)
						contentScroll.SetMinSize(fyne.NewSize(400, 200))

						dialogContent := container.NewVBox(
							widget.NewLabel("Name:"),
							fileNameEntry,
							contentScroll,
						)

						dialogs.ShowCustomConfirm(w, "Edit Devshell", "💾  Save", "Cancel", dialogContent, func(ok bool) {
							if ok {
								runCmd(func() error {
									err := api.UpdateDevshell(shellName, contentEntry.Text)
									if err != nil {
										return err
									}
									if fileNameEntry.Text != shellName {
										err = api.RenameDevShell(shellName, fileNameEntry.Text)
										if err != nil {
											return err
										}
									}
									return nil
								}, "💾  Updating devshell...", false, func() {
									tab.Refresh()
									refreshPendingActions()
								})
							}
						})
					}),
					fyne.NewMenuItem("📋  Duplicate", func() {
						runCmd(func() error {
							return api.DuplicateDevShell(shellName)
						}, "📋  Duplicating devshell...", false, func() {
							tab.Refresh()
						})
					}),
					fyne.NewMenuItem("🗑️  Remove", func() {
						dialogs.ShowConfirm(w, "Remove Devshell", "Are you sure you want to remove "+shellName+"?", func(ok bool) {
							if ok {
								runCmd(func() error {
									return api.RemoveDevshell(shellName)
								}, "🗑️  Removing devshell...", false, func() {
									tab.Refresh()
								})
							}
						})
					}),
				)
				widget.NewPopUpMenu(menu, w.Canvas()).ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(button))
			}
		},
	)

	addShellButton := widget.NewButton("➕  Add Devshell", func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Enter devshell name (optional)")

		contentEntry := widget.NewMultiLineEntry()
		contentEntry.SetPlaceHolder("Enter devshell content here...")
		contentScroll := container.NewScroll(contentEntry)
		contentScroll.SetMinSize(fyne.NewSize(400, 200))

		dialogContent := container.NewVBox(
			widget.NewLabel("Name:"),
			nameEntry,
			contentScroll,
		)

		dialogs.ShowCustomConfirm(w, "Add Devshell", "Add", "Cancel", dialogContent, func(ok bool) {
			if ok {
				runCmd(func() error {
					return api.AddDevshellWithContent(nameEntry.Text, contentEntry.Text)
				}, "➕  Adding devshell...", false, func() {
					tab.Refresh()
				})
			}
		})
	})

	devshellsAccordion := widget.NewAccordion()
	for _, shell := range tab.devshells {
		item := widget.NewAccordionItem(
			shell.Name,
			widget.NewLabel(shell.Description),
		)
		devshellsAccordion.Append(item)
	}

	controls := container.NewVBox(
		addShellButton,
	)
	content := container.NewBorder(controls, nil, nil, nil, tab.list)
	tab.CanvasObject = container.NewPadded(content)
	return tab
}
