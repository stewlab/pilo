package tabs

import (
	"pilo/internal/api"
	"pilo/internal/config"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"pilo/internal/dialogs"
)

type PackagesTab struct {
	fyne.CanvasObject
	refreshInstalled func(showDialog bool)
	refreshCustom    func()
}

func (t *PackagesTab) Refresh() {
	t.refreshInstalled(false) // Do not show dialog on automatic refresh
	t.refreshCustom()
}

func CreatePackagesTab(runCmd func(func() error, string, bool, func()), a fyne.App, w fyne.Window, refreshPendingActions func()) *PackagesTab {
	// showLoading is replaced by dialogs.ShowRunningCommandDialog

	// --- Installed Packages Tab ---
	installedPackagesBinding := binding.NewUntypedList()
	var installedPackagesList *widget.List
	refreshInstalled := func(showDialog bool) {
		go func() {
			getPackages := func() (string, error) {
				pkgs, err := api.GetInstalledPackages()
				if err != nil {
					return "", err
				}
				var items []interface{}
				for _, pkg := range pkgs {
					items = append(items, pkg)
				}
				fyne.Do(func() {
					installedPackagesBinding.Set(items)
					installedPackagesList.Refresh()
				})
				return "Installed packages retrieved successfully!", nil
			}

			if showDialog {
				dialogs.ShowRunningCommandDialog(w, "Getting installed packages...", getPackages, nil)
			} else {
				getPackages()
			}
		}()
	}

	installedPackagesList = widget.NewListWithData(
		installedPackagesBinding,
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, widget.NewLabel("Template"), widget.NewButton("🗑️  Remove", nil))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			untyped, _ := i.(binding.Untyped).Get()
			pkg := untyped.(api.Package)
			label := o.(*fyne.Container).Objects[0].(*widget.Label)
			if !pkg.Installed {
				label.SetText(pkg.Name + " (pending)")
			} else {
				label.SetText(pkg.Name)
			}
			removeButton := o.(*fyne.Container).Objects[1].(*widget.Button)
			removeButton.OnTapped = func() {
				dialogs.ShowConfirm(w, "Remove Package", "Are you sure you want to remove "+pkg.Name+"?", func(ok bool) {
					if ok {
						runCmd(func() error {
							return api.RemovePackage(pkg.Name)
						}, "🗑️  Removing package...", false, func() {
							refreshInstalled(false)
							refreshPendingActions()
						})
					}
				})
			}
		},
	)

	refreshInstalledButton := widget.NewButton("🔄  Refresh", func() {
		refreshInstalled(true)
	})
	installedBox := container.NewBorder(container.NewVBox(widget.NewLabel("Installed Packages"), refreshInstalledButton), nil, nil, nil, installedPackagesList)
	refreshInstalled(false) // Initial load without dialog

	// --- Search packages Tab ---
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Enter search query")
	resultsBinding := binding.NewUntypedList()

	resultsList := widget.NewListWithData(
		resultsBinding,
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, widget.NewLabel("template"), widget.NewButton("📥  Install", nil))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			item, err := i.(binding.Untyped).Get()
			if err != nil {
				fyne.LogError("Error getting from binding", err)
				return
			}
			pkg := item.(api.Package)
			container := o.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			label.SetText(pkg.Name + " - " + pkg.Description)

			installButton := container.Objects[1].(*widget.Button)
			installButton.OnTapped = func() {
				dialogs.ShowConfirm(w, "Install Package", "Are you sure you want to install "+pkg.Name+"?", func(ok bool) {
					if ok {
						runCmd(func() error {
							return api.AddPackage(pkg.Name)
						}, "📥  Adding package...", false, func() {
							refreshInstalled(false)
							refreshPendingActions()
							installButton.Hide()
						})
					}
				})
			}
		},
	)

	sortByPopularityCheck := widget.NewCheck("Sort by popularity", nil)
	freeOnlyCheck := widget.NewCheck("Show only free software", nil)

	searchButton := widget.NewButton("🔍  Search", func() {
		dialogs.ShowRunningCommandDialog(w, "🔍  Searching...", func() (string, error) {
			out, err := api.Search(strings.Fields(searchEntry.Text), sortByPopularityCheck.Checked, freeOnlyCheck.Checked)
			if err != nil {
				config.AddLogEntry("Error searching packages: " + err.Error())
				return "", err
			}
			newItems := make([]interface{}, len(out))
			for i, v := range out {
				newItems[i] = v
			}
			fyne.Do(func() {
				resultsBinding.Set(newItems)
				resultsList.Refresh()
			})
			config.AddLogEntry("Package search completed successfully!")
			return "Search complete!", nil
		}, nil)
	})

	searchControls := container.NewVBox(
		widget.NewLabel("Search Packages"),
		searchEntry,
		sortByPopularityCheck,
		freeOnlyCheck,
		searchButton,
	)
	searchBox := container.NewBorder(searchControls, nil, nil, nil, resultsList)

	// --- Custom Packages Tab ---
	var list *widget.List
	var apps []string
	var err error

	refreshCustom := func() {
		go func() {
			apps, err = api.ListApps()
			if err != nil {
				// Handle error
			}
			if list != nil {
				fyne.Do(func() {
					list.Refresh()
				})
			}
		}()
	}
	refreshCustom()

	list = widget.NewList(
		func() int {
			return len(apps)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
				layout.NewSpacer(),
				widget.NewButton("...", nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			appName := strings.TrimSuffix(apps[i], ".nix")
			hbox := o.(*fyne.Container)
			label := hbox.Objects[0].(*widget.Label)
			label.SetText(appName)
			button := hbox.Objects[2].(*widget.Button)
			button.OnTapped = func() {
				menu := fyne.NewMenu("",
					fyne.NewMenuItem("✏️  Edit", func() {
						content, err := api.GetAppContent(appName)
						if err != nil {
							dialogs.ShowErrorDialog(err, w)
							return
						}
						fileNameEntry := widget.NewEntry()
						fileNameEntry.SetText(appName)

						contentEntry := widget.NewMultiLineEntry()
						contentEntry.SetText(content)
						contentScroll := container.NewScroll(contentEntry)
						contentScroll.SetMinSize(fyne.NewSize(400, 200))

						dialogContent := container.NewVBox(
							widget.NewLabel("Filename:"),
							fileNameEntry,
							contentScroll,
						)

						dialogs.ShowCustomConfirm(w, "Edit Package", "💾  Save", "Cancel", dialogContent, func(ok bool) {
							if ok {
								runCmd(func() error {
									err := api.UpdateApp(appName, contentEntry.Text)
									if err != nil {
										return err
									}
									if fileNameEntry.Text != appName {
										err = api.RenameApp(appName, fileNameEntry.Text)
										if err != nil {
											return err
										}
									}
									return nil
								}, "💾  Updating package...", false, func() {
									refreshCustom()
								})
							}
						})
					}),
					fyne.NewMenuItem("📋  Duplicate", func() {
						runCmd(func() error {
							return api.DuplicateApp(appName)
						}, "📋  Duplicating package...", false, func() {
							refreshCustom()
						})
					}),
					fyne.NewMenuItem("🗑️  Remove", func() {
						dialogs.ShowConfirm(w, "Remove Package", "Are you sure you want to remove "+appName+"?", func(ok bool) {
							if ok {
								runCmd(func() error {
									return api.RemoveApp(appName)
								}, "🗑️  Removing package...", false, func() {
									refreshCustom()
								})
							}
						})
					}))
				widget.NewPopUpMenu(menu, w.Canvas()).ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(button))
			}
		},
	)

	addPackageButton := widget.NewButton("➕  Add Custom Package", func() {
		pnameEntry := widget.NewEntry()
		pnameEntry.SetPlaceHolder("Enter custom package name")
		contentEntry := widget.NewMultiLineEntry()
		contentEntry.SetPlaceHolder("Package definition")
		contentEntry.Wrapping = fyne.TextWrapOff

		getTemplateButton := widget.NewButton("📄  Use Template", func() {
			contentEntry.SetText(api.GetAppTemplate())
		})

		contentScroll := container.NewScroll(contentEntry)
		contentScroll.SetMinSize(fyne.NewSize(400, 200))

		form := container.NewVBox(
			widget.NewLabel("Add Custom Package"),
			pnameEntry,
			contentScroll,
			getTemplateButton,
		)

		dialogs.ShowCustomConfirm(w, "Add Custom Package", "💾  Save", "Cancel", form, func(ok bool) {
			if ok {
				runCmd(func() error {
					return api.AddAppFromContent(pnameEntry.Text, contentEntry.Text)
				}, "➕  Adding custom package...", false, func() {
					refreshCustom()
				})
			}
		})
	})

	addGitPackageButton := widget.NewButton("➕  Add Git Package", func() {
		urlEntry := widget.NewEntry()
		urlEntry.SetPlaceHolder("github:owner/repo or https://github.com/owner/repo.git")

		dialogContent := container.NewVBox(
			widget.NewLabel("Enter git URL"),
			urlEntry,
		)

		dialogs.ShowCustomConfirm(w, "Add Git Package", "💾  Save", "Cancel", dialogContent, func(ok bool) {
			if ok {
				runCmd(func() error {
					return api.AddGitPackage(urlEntry.Text)
				}, "➕  Adding git package...", false, func() {
					refreshInstalled(false)
					refreshPendingActions()
				})
			}
		})
	})

	customPackagesControls := container.NewVBox(
		addPackageButton,
		addGitPackageButton,
		widget.NewSeparator(),
		widget.NewLabel("Existing Custom Packages"),
	)
	customPackagesBox := container.NewBorder(customPackagesControls, nil, nil, nil, list)

	tabs := container.NewAppTabs(
		container.NewTabItem("Search", searchBox),
		container.NewTabItem("Installed", installedBox),
		container.NewTabItem("Custom", customPackagesBox),
	)

	tab := &PackagesTab{
		CanvasObject:     container.NewPadded(tabs),
		refreshInstalled: refreshInstalled,
		refreshCustom:    refreshCustom,
	}
	return tab
}
