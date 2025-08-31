package gui

import (
	"fmt"
	"image/color"
	"os"
	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/dialogs"
	"pilo/internal/gui/tabs"
	"pilo/internal/nix"
	"time" // New import

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var refreshTabs func()

var (
	// Version is the application version, set at build time.
	Version = "0.0.1"
)

func run() {
	a := app.NewWithID("dev.stewlab.pilo")
	config.Init(a)
	a.Settings().SetTheme(&myTheme{})
	w := a.NewWindow("pilo")
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()

	// Handle auto-installation if the config path doesn't exist
	handleAutoInstall(w)

	if nix.GetNixMode() == nix.None {
		dialog.NewInformation("Nix Not Found", "Nix is not installed. Please install it from the preferences tab for full functionality.", w).Show()
	}

	logs := binding.NewStringList()
	logs.Set(config.GetLogs())

	// title := widget.NewLabelWithStyle("pilo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	flakePathEntry := widget.NewEntry()
	flakePathEntry.SetText(config.GetInstallPath())
	flakePathEntry.OnChanged = func(s string) {
		a.Preferences().SetString("installationPath", s)
	}

	pendingActionsBinding := binding.NewStringList()
	gitStatusBinding := binding.NewString()
	gitStatusBinding.Set("Press refresh to see the current git status.")

	var refreshPendingActions func()

	statusButton := widget.NewButton("Checking config status...", func() {
		go func() {
			status, err := api.GetGitStatus(config.GetInstallPath())
			if err != nil {
				fyne.LogError("Failed to get git status", err)
				return
			}
			fyne.Do(func() {
				dialogs.ShowGitStatusDialog(status, w, config.GetInstallPath())
			})
		}()
	})

	runCmd := func(f func() (string, error), msg string, showOutput bool, refresh func()) {
		config.App.Preferences().SetString("currentTime", time.Now().Format(time.RFC3339)) // Set current time for logging
		dialogs.ShowRunningCommandDialog(w, msg, f, func(output string, err error) {
			fyne.Do(func() {
				if err != nil {
					config.AddLogEntry(fmt.Sprintf("Error: %v", err))
				}
				if output != "" {
					config.AddLogEntry(output)
				}
				logs.Set(config.GetLogs()) // Refresh logs after command
				if refresh != nil {
					refresh()
				}
				if refreshPendingActions != nil {
					refreshPendingActions()
				}
			})
		})
	}

	refreshPendingActions = func() {
		go func() {
			actions, err := api.GetPendingActions()
			if err != nil {
				fyne.LogError("Failed to get pending actions", err)
				return
			}
			fyne.Do(func() {
				pendingActionsBinding.Set(actions)
			})

			dirty, err := api.GitStatus(config.GetInstallPath())
			if err != nil {
				fyne.Do(func() {
					statusButton.SetText("Error checking config status.")
					fyne.LogError("Failed to get git status", err)
					gitStatusBinding.Set("Error checking git status.")
				})
				return
			}
			fyne.Do(func() {
				if dirty {
					statusButton.SetText("Uncommitted changes")
					gitStatusBinding.Set("Uncommitted changes.")
				} else {
					statusButton.SetText("No uncommitted changes")
					gitStatusBinding.Set("No uncommitted changes.")
				}
			})
		}()
	}

	appTabs := container.NewAppTabs()

	var refreshableTabs []tabs.Refreshable
	refreshTabs = func() {
		for _, t := range refreshableTabs {
			t.Refresh()
		}
	}

	systemTabContent := tabs.CreateSystemTab(func(f func() (string, error), msg string, showOutput bool, refreshFunc func()) {
		runCmd(f, msg, showOutput, refreshFunc)
	}, config.GetFlakePath(), a.Preferences(), w, refreshPendingActions)
	packagesTabContent := tabs.CreatePackagesTab(func(f func() error, msg string, showOutput bool, refreshFunc func()) {
		runCmd(func() (string, error) {
			err := f()
			return "", err
		}, msg, showOutput, refreshFunc)
	}, a, w, refreshPendingActions)
	devshellsTabContent := tabs.CreateDevshellTab(func(f func() error, msg string, showOutput bool, refreshFunc func()) {
		runCmd(func() (string, error) {
			err := f()
			return "", err
		}, msg, showOutput, refreshFunc)
	}, config.GetFlakePath(), w, refreshPendingActions)
	aliasesTabContent := tabs.CreateAliasesTab(func(f func() error, msg string, showOutput bool, refreshFunc func()) {
		runCmd(func() (string, error) {
			err := f()
			return "", err
		}, msg, showOutput, refreshFunc)
	}, w, refreshPendingActions)
	preferencesTabContent := tabs.CreatePreferencesTab(func(f func() (string, error), msg string, showOutput bool, refreshFunc func()) {
		runCmd(f, msg, showOutput, refreshFunc)
	}, flakePathEntry, a.Preferences(), func(onConfirm func(password string)) {
		dialogs.ShowPasswordDialog(w, onConfirm)
	}, w, appTabs, refreshPendingActions, gitStatusBinding, refreshTabs)

	configEditorTabContent := tabs.CreateConfigEditorTab(w)
	preferencesTab := container.NewTabItem("Preferences", preferencesTabContent.CanvasObject)
	systemTab := container.NewTabItem("System", systemTabContent.CanvasObject)
	packagesTab := container.NewTabItem("Packages", packagesTabContent.CanvasObject)
	devshellsTab := container.NewTabItem("Devshells", devshellsTabContent.CanvasObject)
	aliasesTab := container.NewTabItem("Aliases", aliasesTabContent.CanvasObject)
	configEditorTab := container.NewTabItem("Config Editor", configEditorTabContent.CanvasObject)

	appTabs.SetItems([]*container.TabItem{
		systemTab,
		packagesTab,
		devshellsTab,
		aliasesTab,
		configEditorTab,
		preferencesTab,
	})

	refreshableTabs = append(refreshableTabs, systemTabContent)
	refreshableTabs = append(refreshableTabs, packagesTabContent)
	refreshableTabs = append(refreshableTabs, devshellsTabContent)
	refreshableTabs = append(refreshableTabs, aliasesTabContent)
	refreshableTabs = append(refreshableTabs, preferencesTabContent)
	refreshableTabs = append(refreshableTabs, configEditorTabContent)

	// Set tab location from preferences
	tabLocationStr := a.Preferences().StringWithFallback("tabPosition", "Leading")
	var tabLocation container.TabLocation
	switch tabLocationStr {
	case "Top":
		tabLocation = container.TabLocationTop
	case "Bottom":
		tabLocation = container.TabLocationBottom
	case "Trailing":
		tabLocation = container.TabLocationTrailing
	default:
		tabLocation = container.TabLocationLeading
	}
	appTabs.SetTabLocation(tabLocation)

	logsButton := widget.NewButton("📜 Logs", func() {
		dialogs.ShowLogsDialog(logs, w)
	})

	versionLabel := widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignLeading, fyne.TextStyle{})
	statusBar := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(
			container.New(
				&statusBarLayout{},
				logsButton,
				statusButton,
				versionLabel,
			),
		),
	)

	content := container.NewBorder(
		nil,
		statusBar,
		nil,
		nil,
		appTabs,
	)

	w.SetContent(content)
	refreshTabs() // Call refresh to initialize all tabs
	w.ShowAndRun()
}

// Run is the public entry point for the GUI.
func Run(version string) {
	Version = version
	run()
}

func Refresh() {
	if refreshTabs != nil {
		refreshTabs()
	}
}

func handleAutoInstall(w fyne.Window) {
	installPath := config.GetInstallPath()
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		dialog.NewInformation(
			"Welcome to Pilo",
			"Pilo configuration not found. The default configuration will be installed automatically.",
			w,
		).Show()
		go func() {
			err := api.AutoInstall(installPath, w)
			fyne.Do(func() {
				if err != nil {
					dialog.NewError(fmt.Errorf("failed to install Pilo configuration: %w", err), w).Show()
				} else {
					dialog.NewInformation("Installation Complete", "Pilo has been installed successfully.", w).Show()
					w.Content().Refresh()
				}
			})
		}()
	}
}

type statusBarLayout struct{}

func (s *statusBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	logsButton := objects[0]
	statusButton := objects[1]
	versionLabel := objects[2]

	// Logs button on the far right
	logsButton.Resize(logsButton.MinSize())
	logsButton.Move(fyne.NewPos(size.Width-logsButton.MinSize().Width, (size.Height-logsButton.MinSize().Height)/2))

	// Version label on the far left
	versionLabel.Resize(versionLabel.MinSize())
	versionLabel.Move(fyne.NewPos(0, (size.Height-versionLabel.MinSize().Height)/2))

	// Status button takes up the remaining space in the middle
	statusButton.Resize(fyne.NewSize(size.Width-logsButton.MinSize().Width-versionLabel.MinSize().Width-theme.Padding()*2, statusButton.MinSize().Height))
	statusButton.Move(fyne.NewPos(versionLabel.MinSize().Width+theme.Padding(), (size.Height-statusButton.MinSize().Height)/2))
}

func (s *statusBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minHeight := float32(0)
	for _, o := range objects {
		if o.MinSize().Height > minHeight {
			minHeight = o.MinSize().Height
		}
	}
	return fyne.NewSize(0, minHeight)
}

var _ fyne.Theme = (*myTheme)(nil)

type myTheme struct{}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
