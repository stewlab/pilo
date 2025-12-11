package tabs

import (
	"fmt"
	"pilo/internal/api"
	"pilo/internal/dialogs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DevshellTab struct {
	fyne.CanvasObject
	templateList  *widget.List
	userList      *widget.List
	templates     []api.Devshell
	userDevshells []api.Devshell
	runCmd        func(func() error, string, bool, func())
	window        fyne.Window
}

func (t *DevshellTab) Refresh() {
	var err error
	t.templates, err = api.ListDevshellTemplates()
	if err != nil {
		fyne.LogError("Failed to list devshell templates", err)
	}
	t.templateList.Refresh()

	t.userDevshells, err = api.ListUserDevshells()
	if err != nil {
		fyne.LogError("Failed to list user devshells", err)
	}
	t.userList.Refresh()
}

func CreateDevshellTab(runCmd func(func() error, string, bool, func()), w fyne.Window) *DevshellTab {
	tab := &DevshellTab{
		runCmd: runCmd,
		window: w,
	}

	tab.templateList = widget.NewList(
		func() int {
			return len(tab.templates)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewButton("✨ Initialize", nil),
				widget.NewLabel("Template Name"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			template := tab.templates[i]
			c := o.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			label.SetText(template.Name)

			button := c.Objects[1].(*widget.Button)
			button.OnTapped = func() {
				dirEntry := widget.NewEntry()
				dirEntry.SetPlaceHolder("my-new-project")

				dialogs.ShowCustomConfirm(w, "Initialize Devshell", "Initialize", "Cancel",
					container.NewVBox(
						widget.NewLabel("Enter a directory name for the new devshell:"),
						dirEntry,
					),
					func(ok bool) {
						if !ok || dirEntry.Text == "" {
							return
						}
			       runCmd(func() error {
				       err := api.InitDevshellFromTemplate(template.Name, dirEntry.Text)
				       if err == nil {
					       tab.Refresh()
				       }
				       return err
			       }, "✨ Initializing devshell...", true, tab.Refresh)
					},
				)
			}
		},
	)

	tab.userList = widget.NewList(
		func() int {
			return len(tab.userDevshells)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				container.NewHBox(
					widget.NewButton("✏️ Edit", nil),
					widget.NewButton("🚀 Enter", nil),
					widget.NewButton("📋 Duplicate", nil),
					widget.NewButton("🗑️ Delete", nil),
				),
				widget.NewLabel("Devshell Name"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			devshell := tab.userDevshells[i]
			c := o.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			label.SetText(devshell.Name)

			buttons := c.Objects[1].(*fyne.Container)
			editButton := buttons.Objects[0].(*widget.Button)
			enterButton := buttons.Objects[1].(*widget.Button)
			duplicateButton := buttons.Objects[2].(*widget.Button)
			deleteButton := buttons.Objects[3].(*widget.Button)

			editButton.OnTapped = func() {
				runCmd(func() error {
					return api.EditDevshell(devshell.Path)
				}, "Opening editor...", false, nil)
			}

			enterButton.OnTapped = func() {
				runCmd(func() error {
					return api.EnterDevshell(devshell.Path)
				}, "Entering devshell...", false, nil)
			}

			duplicateButton.OnTapped = func() {
				nameEntry := widget.NewEntry()
				nameEntry.SetPlaceHolder(devshell.Name + "-copy")

				dialogs.ShowCustomConfirm(w, "Duplicate Devshell", "Duplicate", "Cancel",
					container.NewVBox(
						widget.NewLabel("Enter a new name for the duplicated devshell:"),
						nameEntry,
					),
					func(ok bool) {
						if !ok || nameEntry.Text == "" {
							return
						}
						runCmd(func() error {
							return api.DuplicateDevshell(devshell.Path, nameEntry.Text)
						}, "Duplicating devshell...", true, tab.Refresh)
					},
				)
			}

			deleteButton.OnTapped = func() {
				dialogs.ShowCustomConfirm(w, "Delete Devshell", "Delete", "Cancel",
					widget.NewLabel(fmt.Sprintf("Are you sure you want to delete devshell '%s'?", devshell.Name)),
					func(ok bool) {
						if !ok {
							return
						}
						runCmd(func() error {
							return api.DeleteDevshell(devshell.Path)
						}, "Deleting devshell...", true, tab.Refresh)
					},
				)
			}
		},
	)
	tab.Refresh() // Initial load

	tab.CanvasObject = container.NewBorder(
		container.NewPadded(widget.NewLabel("Devshell Templates:")),
		nil,
		nil,
		nil,
		container.NewVSplit(
			container.NewPadded(tab.templateList),
			container.NewBorder(
				container.NewPadded(widget.NewLabel("Your Devshells:")),
				nil,
				nil,
				nil,
				container.NewPadded(tab.userList),
			),
		),
	)
	return tab
}
