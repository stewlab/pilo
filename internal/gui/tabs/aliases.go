package tabs

import (
	"pilo/internal/api"
	"pilo/internal/dialogs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type AliasesTab struct {
	fyne.CanvasObject
	refreshAliases func()
}

func (t *AliasesTab) Refresh() {
	t.refreshAliases()
}

func CreateAliasesTab(runCmd func(func() error, string, bool, func()), w fyne.Window, refreshPendingActions func()) *AliasesTab {
	aliasBinding := binding.NewUntypedList()
	refreshAliases := func() {
		aliases, err := api.GetAliases()
		if err != nil {
			// Handle error
			return
		}
		// Convert map to a slice of items for the binding
		newItems := make([]interface{}, 0, len(aliases))
		for name, cmd := range aliases {
			newItems = append(newItems, map[string]string{"name": name, "cmd": cmd})
		}
		aliasBinding.Set(newItems)
	}

	refreshAliases() // Initial load

	aliasList := widget.NewListWithData(
		aliasBinding,
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("template"),
				widget.NewLabel("template"),
				layout.NewSpacer(),
				widget.NewButton("...", nil),
			)
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			item, err := i.(binding.Untyped).Get()
			if err != nil {
				fyne.LogError("Error getting from binding", err)
				return
			}
			alias := item.(map[string]string)
			name := alias["name"]
			command := alias["cmd"]
			hbox := o.(*fyne.Container)
			hbox.Objects[0].(*widget.Label).SetText(name)
			hbox.Objects[1].(*widget.Label).SetText(command)

			button := hbox.Objects[3].(*widget.Button)
			button.OnTapped = func() {
				menu := fyne.NewMenu("",
					fyne.NewMenuItem("✏️  Edit", func() {
						editNameEntry := widget.NewEntry()
						editNameEntry.SetText(name)
						editCommandEntry := widget.NewEntry()
						editCommandEntry.SetText(command)

						dialogs.ShowForm(w, "Edit Alias", "💾  Save", "Cancel", []*widget.FormItem{
							widget.NewFormItem("Name", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, editNameEntry.MinSize().Height)), editNameEntry)),
							widget.NewFormItem("Command", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, editCommandEntry.MinSize().Height)), editCommandEntry)),
						}, func(b bool) {
							if !b {
								return
							}
							runCmd(func() error {
								return api.UpdateAlias(name, editNameEntry.Text, editCommandEntry.Text)
							}, "💾  Updating alias...", false, func() {
								refreshAliases()
								refreshPendingActions()
							})
						})
					}),
					fyne.NewMenuItem("📋  Duplicate", func() {
						runCmd(func() error {
							return api.DuplicateAlias(name, command)
						}, "📋  Duplicating alias...", false, func() {
							refreshAliases()
						})
					}),
					fyne.NewMenuItem("🗑️  Remove", func() {
						dialogs.ShowConfirm(w, "Remove Alias", "Are you sure you want to remove alias "+name+"?", func(ok bool) {
							if ok {
								runCmd(func() error {
									return api.RemoveAlias(name)
								}, "🗑️  Removing alias...", false, func() {
									refreshAliases()
								})
							}
						})
					}))
				widget.NewPopUpMenu(menu, w.Canvas()).ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(button))
			}
		},
	)

	addButton := widget.NewButton("➕  Add Alias", func() {
		addNameEntry := widget.NewEntry()
		addNameEntry.SetPlaceHolder("Enter alias name")
		addCommandEntry := widget.NewEntry()
		addCommandEntry.SetPlaceHolder("Enter command")

		dialogs.ShowForm(w, "Add Alias", "Add", "Cancel", []*widget.FormItem{
			widget.NewFormItem("Name", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, addNameEntry.MinSize().Height)), addNameEntry)),
			widget.NewFormItem("Command", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, addCommandEntry.MinSize().Height)), addCommandEntry)),
		}, func(b bool) {
			if !b {
				return
			}
			runCmd(func() error {
				return api.AddAlias(addNameEntry.Text, addCommandEntry.Text)
			}, "➕  Adding alias...", false, func() {
				refreshAliases()
			})
		})
	})

	content := container.NewBorder(
		container.NewVBox(
			addButton,
		),
		nil, nil, nil,
		aliasList,
	)
	tab := &AliasesTab{
		CanvasObject:   container.NewPadded(content),
		refreshAliases: refreshAliases,
	}
	return tab
}
