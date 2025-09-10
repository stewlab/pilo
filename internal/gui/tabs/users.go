package tabs

import (
	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/dialogs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type UsersTab struct {
	fyne.CanvasObject
	refreshUsers func()
}

func (t *UsersTab) Refresh() {
	t.refreshUsers()
}

func CreateUsersTab(runCmd func(func() error, string, bool, func()), window fyne.Window, refreshPendingActions func()) *UsersTab {
	userBinding := binding.NewUntypedList()
	refreshUsers := func() {
		users, err := api.GetUsers()
		if err != nil {
			// Handle error
			return
		}
		newItems := make([]interface{}, len(users))
		for i, u := range users {
			newItems[i] = u
		}
		userBinding.Set(newItems)
	}

	refreshUsers() // Initial load

	userList := widget.NewListWithData(
		userBinding,
		func() fyne.CanvasObject {
			return container.NewHBox(
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
			user := item.(config.User)
			hbox := o.(*fyne.Container)
			hbox.Objects[0].(*widget.Label).SetText(user.Username + " (" + user.Name + ") <" + user.Email + ">")

			button := hbox.Objects[2].(*widget.Button)
			button.OnTapped = func() {
				menu := fyne.NewMenu("",
					fyne.NewMenuItem("✏️  Edit", func() {
						showUserDialog(window, &user, runCmd, refreshUsers)
					}),
					fyne.NewMenuItem("🗑️  Remove", func() {
						dialogs.ShowConfirm(window, "Remove User", "Are you sure you want to remove user "+user.Username+"?", func(ok bool) {
							if ok {
								runCmd(func() error {
									return api.RemoveUser(user.Username)
								}, "🗑️  Removing user...", false, func() {
									refreshUsers()
								})
							}
						})
					}))
				widget.NewPopUpMenu(menu, window.Canvas()).ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(button))
			}
		},
	)

	addButton := widget.NewButton("➕  Add User", func() {
		showUserDialog(window, nil, runCmd, refreshUsers)
	})

	content := container.NewBorder(
		container.NewVBox(
			addButton,
		),
		nil, nil, nil,
		userList,
	)
	tab := &UsersTab{
		CanvasObject: container.NewPadded(content),
		refreshUsers: refreshUsers,
	}
	return tab
}

func showUserDialog(window fyne.Window, user *config.User, runCmd func(func() error, string, bool, func()), refresh func()) {
	usernameEntry := widget.NewEntry()
	nameEntry := widget.NewEntry()
	emailEntry := widget.NewEntry()

	var oldUsername string
	if user != nil {
		usernameEntry.SetText(user.Username)
		nameEntry.SetText(user.Name)
		emailEntry.SetText(user.Email)
		oldUsername = user.Username
	}

	dialogs.ShowForm(window, "User", "💾  Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Username", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, usernameEntry.MinSize().Height)), usernameEntry)),
		widget.NewFormItem("Name", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, nameEntry.MinSize().Height)), nameEntry)),
		widget.NewFormItem("Email", container.New(layout.NewGridWrapLayout(fyne.NewSize(300, emailEntry.MinSize().Height)), emailEntry)),
	}, func(b bool) {
		if !b {
			return
		}
		if user == nil { // Add
			runCmd(func() error {
				return api.AddUser(usernameEntry.Text, nameEntry.Text, emailEntry.Text)
			}, "➕  Adding user...", false, func() {
				refresh()
			})
		} else { // Update
			runCmd(func() error {
				return api.UpdateUser(oldUsername, usernameEntry.Text, nameEntry.Text, emailEntry.Text)
			}, "💾  Updating user...", false, func() {
				refresh()
			})
		}
	})
}
