package tabs

import (
	"fmt"
	"pilo/internal/api"
	"pilo/internal/config" // New import
	"pilo/internal/gui/components"
	"pilo/internal/nix"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type PreferencesTab struct {
	CanvasObject fyne.CanvasObject

	installPathEntry    *components.SafeEntry
	registryNameEntry   *components.SafeEntry
	remoteUrlEntry      *components.SafeEntry
	remoteBranchEntry   *components.SafeEntry
	pushOnCommitCheck   *widget.Check
	systemEntry         *components.SafeEntry
	usernameEntry       *components.SafeEntry
	nixpkgsEntry        *components.SafeEntry
	homeManagerEntry    *components.SafeEntry
	nixInstallCmdEntry  *components.SafeEntry
	logRetentionEntry   *components.SafeEntry
	customTerminalEntry *components.SafeEntry

	systemActionsCheck *widget.Check
	appActionsCheck    *widget.Check
	pkgActionsCheck    *widget.Check
}

func (t *PreferencesTab) Refresh() {
	t.installPathEntry.SetText(config.GetInstallPath())
	t.registryNameEntry.SetText(config.GetRegistryName())
	if remoteURL, err := config.GetRemoteUrl(); err == nil {
		t.remoteUrlEntry.SetText(remoteURL)
	}
	if remoteBranch, err := config.GetRemoteBranch(); err == nil {
		t.remoteBranchEntry.SetText(remoteBranch)
	}
	if pushOnCommit, err := config.GetPushOnCommit(); err == nil {
		t.pushOnCommitCheck.SetChecked(pushOnCommit)
	}
	if system, err := config.GetSystem(); err == nil {
		t.systemEntry.SetText(system.Type)
	}
	if username, err := config.GetUsername(); err == nil {
		t.usernameEntry.SetText(username)
	}
	t.nixpkgsEntry.SetText(config.GetNixpkgsUrl())
	t.homeManagerEntry.SetText(config.GetHomeManagerUrl())
	t.nixInstallCmdEntry.SetText(config.GetNixInstallCmd())
	t.logRetentionEntry.SetText(string(rune(config.GetLogHistoryRetention())))
	t.customTerminalEntry.SetText(config.GetCustomTerminal())

	// Refresh commit triggers
	triggers, err := config.GetCommitTriggers()
	if err == nil {
		t.systemActionsCheck.SetChecked(contains(triggers, "rebuild"))
		t.appActionsCheck.SetChecked(contains(triggers, "add_app"))
		t.pkgActionsCheck.SetChecked(contains(triggers, "add_pkg"))
	}
}

// CreatePreferencesTab creates the content for the "Preferences" tab
func CreatePreferencesTab(
	runCmd func(f func() (string, error), msg string, showOutput bool, refresh func()),
	flakePathEntry *widget.Entry,
	prefs fyne.Preferences,
	showPasswordDialog func(onConfirm func(password string)),
	w fyne.Window,
	appTabs *container.AppTabs,
	refreshPendingActions func(),
	gitStatusBinding binding.String,
	refreshAllTabs func(),
) *PreferencesTab {
	tab := &PreferencesTab{}

	// Install management flake
	tab.installPathEntry = components.NewSafeEntry()
	tab.installPathEntry.SetText(config.GetInstallPath())
	tab.installPathEntry.OnChanged = func(s string) {
		prefs.SetString("installationPath", s)
		flakePathEntry.SetText(s)
	}

	registryName := prefs.StringWithFallback("registryName", "pilo")
	tab.registryNameEntry = components.NewSafeEntry()
	tab.registryNameEntry.SetText(registryName)
	tab.registryNameEntry.OnChanged = func(s string) {
		prefs.SetString("registryName", s)
	}

	// Remote Git Repository
	tab.remoteUrlEntry = components.NewSafeEntry()
	tab.remoteUrlEntry.OnChanged = func(s string) {
		config.SetRemoteUrl(s)
	}

	tab.remoteBranchEntry = components.NewSafeEntry()
	tab.remoteBranchEntry.OnChanged = func(s string) {
		config.SetRemoteBranch(s)
	}

	writeAccessWarning := widget.NewLabelWithStyle("Requires write access to the Remote Git URL.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true})
	writeAccessWarning.Wrapping = fyne.TextWrapWord
	writeAccessWarning.Hide()

	tab.pushOnCommitCheck = widget.NewCheck("Push on Commit", func(b bool) {
		config.SetPushOnCommit(b)
		if b {
			writeAccessWarning.Show()
		} else {
			writeAccessWarning.Hide()
		}
	})

	// System and Username
	tab.systemEntry = components.NewSafeEntry()
	tab.systemEntry.OnChanged = func(s string) {
		system, err := config.GetSystem()
		if err != nil {
			// Handle error, maybe show a dialog or log it
			return
		}
		system.Type = s
		config.SetSystem(system)
	}

	tab.usernameEntry = components.NewSafeEntry()
	tab.usernameEntry.OnChanged = func(s string) {
		config.SetUsername(s)
	}

	// Flake overrides
	tab.nixpkgsEntry = components.NewSafeEntry()
	tab.nixpkgsEntry.OnChanged = func(s string) {
		config.SetNixpkgsUrl(s)
	}

	tab.homeManagerEntry = components.NewSafeEntry()
	tab.homeManagerEntry.OnChanged = func(s string) {
		config.SetHomeManagerUrl(s)
	}

	tab.nixInstallCmdEntry = components.NewSafeEntry()
	tab.nixInstallCmdEntry.OnChanged = func(s string) {
		config.SetNixInstallCmd(s)
	}

	tab.logRetentionEntry = components.NewSafeEntry()
	tab.logRetentionEntry.OnChanged = func(s string) {
		if i, err := strconv.Atoi(s); err == nil {
			config.SetLogHistoryRetention(i)
		}
	}

	tab.systemActionsCheck = widget.NewCheck("System Actions (rebuild)", nil)
	tab.appActionsCheck = widget.NewCheck("Application Actions (add/remove)", nil)
	tab.pkgActionsCheck = widget.NewCheck("Package Actions (add/remove)", nil)

	tab.customTerminalEntry = components.NewSafeEntry()
	tab.customTerminalEntry.OnChanged = func(s string) {
		config.SetCustomTerminal(s)
	}

	reinstallButton := widget.NewButton("🚀  Reinstall Pilo Config", func() {
		reinstallDialog := dialog.NewCustomConfirm(
			"Reinstall Pilo Configuration",
			"Preserve Settings",
			"Reset to Defaults",
			widget.NewLabel("This will commit any uncommitted changes and then reinstall the Pilo configuration. You can choose to preserve your existing settings or reset them to the default values."),
			func(preserve bool) {
				if !preserve {
					dialog.NewConfirm(
						"Confirm Reset",
						"Are you sure you want to reset all settings to their default values? This action cannot be undone.",
						func(confirm bool) {
							if confirm {
								reinstall(true, runCmd, refreshPendingActions, refreshAllTabs)
							}
						},
						w,
					).Show()
				} else {
					reinstall(false, runCmd, refreshPendingActions, refreshAllTabs)
				}
			},
			w,
		)
		reinstallDialog.Show()
	})
	reinstallButton.Importance = widget.HighImportance

	// Set initial state for pushOnCommitCheck and warning
	pushOnCommit, _ := config.GetPushOnCommit()
	tab.pushOnCommitCheck.SetChecked(pushOnCommit)
	if pushOnCommit {
		writeAccessWarning.Show()
	}

	// Load initial system and username
	if system, err := config.GetSystem(); err == nil {
		tab.systemEntry.SetText(system.Type)
	}
	if username, err := config.GetUsername(); err == nil {
		tab.usernameEntry.SetText(username)
	}
	// Load initial remote URL and branch
	if remoteURL, err := config.GetRemoteUrl(); err == nil {
		tab.remoteUrlEntry.SetText(remoteURL)
	}
	if remoteBranch, err := config.GetRemoteBranch(); err == nil {
		tab.remoteBranchEntry.SetText(remoteBranch)
	}

	installForm := widget.NewForm(
		widget.NewFormItem("Installation Path", tab.installPathEntry),
		widget.NewFormItem("Nix Registry Name", tab.registryNameEntry),
		widget.NewFormItem("Remote Git URL", tab.remoteUrlEntry),
		widget.NewFormItem("Remote Git Branch", tab.remoteBranchEntry),
		widget.NewFormItem("Push on Commit", tab.pushOnCommitCheck),
		widget.NewFormItem("", writeAccessWarning),
	)

	loggingForm := widget.NewForm(
		widget.NewFormItem("Log History Retention", tab.logRetentionEntry),
	)

	overrideForm := widget.NewForm(
		widget.NewFormItem("Nixpkgs URL", tab.nixpkgsEntry),
		widget.NewFormItem("Home Manager URL", tab.homeManagerEntry),
		widget.NewFormItem("System", tab.systemEntry),
		widget.NewFormItem("Username", tab.usernameEntry),
	)

	nixInstallButton := widget.NewButton("📥  Install Nix", func() {
		runCmd(func() (string, error) { // Change signature to return (string, error)
			err := api.EnsureNixInstalled()
			if err != nil {
				config.AddLogEntry("Error installing Nix: " + err.Error())
				return "", err
			}
			config.AddLogEntry("Nix installed successfully!")
			return "Nix installed successfully!", nil
		}, "Installing Nix", true, nil) // Add msg and showOutput parameters
	})
	if nix.GetNixMode() == nix.NixOS {
		nixInstallButton.Disable()
	}

	// Commit Triggers
	updateCommitTriggers := func() {
		var triggers []string
		if tab.systemActionsCheck.Checked {
			triggers = append(triggers, "rebuild")
		}
		if tab.appActionsCheck.Checked {
			triggers = append(triggers, "add_app", "remove_app")
		}
		if tab.pkgActionsCheck.Checked {
			triggers = append(triggers, "add_pkg", "remove_pkg")
		}
		if err := config.SetCommitTriggers(triggers); err != nil {
			// Handle error, maybe show a dialog or log it
			config.AddLogEntry(fmt.Sprintf("Error setting commit triggers: %v", err))
		}
	}

	tab.systemActionsCheck.OnChanged = func(b bool) { updateCommitTriggers() }
	tab.appActionsCheck.OnChanged = func(b bool) { updateCommitTriggers() }
	tab.pkgActionsCheck.OnChanged = func(b bool) { updateCommitTriggers() }

	commitTriggersForm := widget.NewForm(
		widget.NewFormItem("System Actions", tab.systemActionsCheck),
		widget.NewFormItem("Application Actions", tab.appActionsCheck),
		widget.NewFormItem("Package Actions", tab.pkgActionsCheck),
	)
	commitTriggersContainer := container.NewVBox(
		newWrappingLabel("Choose which actions trigger an automatic git commit of your configuration."),
		commitTriggersForm,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Pilo Configuration Management", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("This allows pilo to manage its own configuration as a flake, enabling versioning and rollbacks."),
		installForm,
		container.NewPadded(container.NewHBox(layout.NewSpacer(), reinstallButton,
			widget.NewButton("📥  Restore from Remote", func() {
				dialog.NewConfirm(
					"Confirm Restore",
					"This will overwrite your local configuration with the one from the remote repository. This is a destructive action and cannot be undone. Are you sure you want to proceed?",
					func(confirm bool) {
						if confirm {
							runCmd(func() (string, error) {
								path := config.GetInstallPath()
								remoteURL, _ := config.GetRemoteUrl()
								if remoteURL == "" {
									return "Remote URL is not set.", nil
								}
								branch, _ := config.GetRemoteBranch()
								err := api.GitRestore(path, remoteURL, branch, nil, "")
								if err != nil {
									if err == api.ErrDirtyRepository {
										// If dirty, create a backup first
										backupErr := api.GitBackup(path)
										if backupErr != nil {
											return "", fmt.Errorf("failed to create backup: %w", backupErr)
										}
										config.AddLogEntry("Created backup of dirty repository before restoring.")

										// Now, restore (discarding local changes as they are backed up)
										strategy := api.GitRestoreDiscard
										err = api.GitRestore(path, remoteURL, branch, &strategy, "")
										if err != nil {
											return "", err
										}
									} else {
										return "", err
									}
								}

								refreshPendingActions()
								return "Configuration restored successfully!", nil
							}, "Restoring Configuration", true, nil)
						}
					},
					w,
				).Show()
			}),
		)),
		widget.NewAccordion(
			&widget.AccordionItem{
				Title: "Automatic Git Commits",
				Detail: container.NewVBox(
					commitTriggersContainer,
				),
			},
		),
		widget.NewLabelWithStyle("Manual Git Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("Manually commit your current configuration changes with a custom message, or create a backup."),
		container.NewPadded(container.NewHBox(layout.NewSpacer(), widget.NewButton("💾 Backup Config Locally Now", func() {
			runCmd(func() (string, error) {
				err := api.GitBackup(config.GetInstallPath())
				if err != nil {
					return "", err
				}
				return "Backup created successfully!", nil
			}, "💾 Creating Backup", true, nil)
		}),
			widget.NewButton("🚀 Sync with Remote", func() {
				runCmd(func() (string, error) {
					err := api.GitSync(config.GetInstallPath())
					if err != nil {
						return "", err
					}
					return "Successfully synced with remote!", nil
				}, " syncing with remote", true, nil)
			}),
		)),
		widget.NewLabelWithStyle("Git Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		func() fyne.CanvasObject {
			statusLabel := widget.NewLabelWithData(gitStatusBinding)
			refreshButton := widget.NewButton("🔄 Refresh", func() {
				refreshPendingActions()
			})
			return container.NewVBox(
				statusLabel,
				container.NewHBox(layout.NewSpacer(), refreshButton),
			)
		}(),
		widget.NewLabelWithStyle("Nix Package Manager", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("If Nix is not installed, you can use this to install it. The command used can be customized."),
		widget.NewForm(
			widget.NewFormItem("Install Command", tab.nixInstallCmdEntry),
		),
		container.NewHBox(layout.NewSpacer(), nixInstallButton),
		widget.NewLabelWithStyle("Flake Overrides", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("Override the default URLs for nixpkgs and home-manager. This is useful for pinning to specific versions or using forks."),
		overrideForm,
		widget.NewLabelWithStyle("Logging", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("Set the number of days to retain log history."),
		loggingForm,
		widget.NewLabelWithStyle("Appearance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		newWrappingLabel("Change the appearance of the application."),
		widget.NewForm(
			widget.NewFormItem("Tab Position", func() fyne.CanvasObject {
				tabPosition := widget.NewSelect([]string{"Leading", "Top", "Bottom", "Trailing"}, func(s string) {
					prefs.SetString("tabPosition", s)
					var tabLocation container.TabLocation
					switch s {
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
				})
				tabPosition.SetSelected(prefs.StringWithFallback("tabPosition", "Leading"))
				return tabPosition
			}()),
		),
		widget.NewForm(
			widget.NewFormItem("Reset Pilo", widget.NewButton("Reset", func() {
				dialog.NewConfirm(
					"Clear Logs",
					"Are you sure you want to clear all logs? This cannot be undone.",
					func(confirm bool) {
						if confirm {
							if err := api.Reset(); err != nil {
								dialog.NewError(err, w).Show()
								return
							}
							dialog.NewInformation("Logs Cleared", "All logs have been cleared.", w).Show()
						}
					},
					w,
				).Show()
			})),
		),
	)

	devshellForm := widget.NewForm(
		widget.NewFormItem("Custom Terminal Command", tab.customTerminalEntry),
	)

	content.Add(widget.NewLabelWithStyle("Devshells", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	content.Add(newWrappingLabel("Set a custom terminal command to use when entering a devshell."))
	content.Add(devshellForm)

	tab.CanvasObject = container.NewScroll(
		container.NewPadded(content),
	)
	return tab
}

type wrappingLabel struct {
	widget.Label
}

func newWrappingLabel(text string) *wrappingLabel {
	label := &wrappingLabel{}
	label.ExtendBaseWidget(label)
	label.SetText(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func reinstall(
	reset bool,
	runCmd func(f func() (string, error), msg string, showOutput bool, refresh func()),
	refreshPendingActions func(),
	refreshAllTabs func(),
) {
	runCmd(func() (string, error) {
		path := config.GetInstallPath()
		registry := config.GetRegistryName()
		var existingConfig *config.BaseConfig
		var err error

		if !reset {
			existingConfig, err = config.ReadConfig()
			if err != nil {
				return "", fmt.Errorf("error reading existing config: %w", err)
			}
		}

		// Add and commit changes before reinstalling
		if err := api.GitAdd(path); err != nil {
			return "", fmt.Errorf("error adding changes: %w", err)
		}
		if err := api.GitCommit(path, "pilo: pre-reinstall commit"); err != nil {
			return "", fmt.Errorf("error committing changes: %w", err)
		}

		remoteURL := ""
		if !reset && existingConfig != nil {
			remoteURL = existingConfig.RemoteURL
		}

		if err := api.Inflate(path, remoteURL, true); err != nil {
			return "", err
		}

		if !reset {
			// Read the new config and merge the old settings
			newConfig, err := config.ReadConfig()
			if err != nil {
				return "", fmt.Errorf("error reading new config: %w", err)
			}
			newConfig.RemoteURL = existingConfig.RemoteURL
			newConfig.RemoteBranch = existingConfig.RemoteBranch
			newConfig.PushOnCommit = existingConfig.PushOnCommit
			newConfig.System = existingConfig.System
			newConfig.Packages = existingConfig.Packages
			newConfig.Aliases = existingConfig.Aliases
			if err := config.WriteConfig(newConfig); err != nil {
				return "", fmt.Errorf("error writing merged config: %w", err)
			}
		}

		if nix.GetNixMode() == nix.NixOS && remoteURL == "" {
			if err := api.CopyNixOSConfigs(path); err != nil {
				return "", err
			}
		}

		err = api.InstallConfig(path, registry, "")
		if err != nil {
			return "", err
		}
		refreshPendingActions()
		return "Pilo configuration reinstalled successfully!", nil
	}, "Reinstalling Pilo Configuration", true, func() {
		refreshAllTabs()
	})
}
