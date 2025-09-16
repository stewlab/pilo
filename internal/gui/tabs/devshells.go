package tabs

import (
	"pilo/internal/api"
	"pilo/internal/dialogs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

func CreateDevshellTab(runCmd func(func() error, string, bool, func()), w fyne.Window) *DevshellTab {
	tab := &DevshellTab{}

	var err error
	tab.devshells, err = api.ListDevshells()
	if err != nil {
		fyne.LogError("Failed to list devshell templates", err)
	}

	tab.list = widget.NewList(
		func() int {
			return len(tab.devshells)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewButton("✨ Initialize", nil),
				widget.NewLabel("Template Name"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			template := tab.devshells[i]
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
							return api.InitDevshellFromTemplate(template.Name, dirEntry.Text)
						}, "✨ Initializing devshell...", true, nil)
					},
				)
			}
		},
	)

	content := container.NewPadded(tab.list)
	tab.CanvasObject = content
	return tab
}
