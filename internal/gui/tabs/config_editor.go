package tabs

import (
	"os"
	"path/filepath"

	"pilo/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ConfigEditorTab struct {
	CanvasObject fyne.CanvasObject
	tree         *widget.Tree
	editor       *widget.Entry
	selectedFile string
	saveButton   *widget.Button
}

func (t *ConfigEditorTab) Refresh() {
	t.selectedFile = ""
	t.editor.SetText("")
	t.saveButton.Disable()
	if t.tree != nil {
		t.tree.UnselectAll()
		t.tree.Refresh()
	}
}

func CreateConfigEditorTab(win fyne.Window) *ConfigEditorTab {
	tab := &ConfigEditorTab{}
	tab.editor = widget.NewMultiLineEntry()
	tab.editor.Wrapping = fyne.TextWrapOff

	tab.saveButton = widget.NewButton("Save", func() {
		if tab.selectedFile == "" {
			return
		}
		err := os.WriteFile(tab.selectedFile, []byte(tab.editor.Text), 0644)
		if err != nil {
			dialog.ShowError(err, win)
		} else {
			dialog.ShowInformation("Success", "File saved successfully", win)
		}
	})
	tab.saveButton.Disable()

	// Get the root path
	rootPath := config.GetFlakePath()

	// Add refresh button for the tree
	refreshButton := widget.NewButton("Refresh", func() {
		tab.Refresh()
	})

	tab.tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			var path string

			// Handle root case
			if uid == "" {
				path = rootPath
			} else {
				path = uid
			}

			// Verify the path exists
			if _, err := os.Stat(path); err != nil {
				fyne.LogError("Path does not exist: "+path, err)
				return []widget.TreeNodeID{}
			}

			// Read directory contents
			entries, err := os.ReadDir(path)
			if err != nil {
				fyne.LogError("Failed to read directory: "+path, err)
				return []widget.TreeNodeID{}
			}

			var children []widget.TreeNodeID
			for _, entry := range entries {
				// Skip hidden files/directories (starting with .)
				if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
					continue
				}
				childPath := filepath.Join(path, entry.Name())
				children = append(children, childPath)
			}
			return children
		},
		func(uid widget.TreeNodeID) bool {
			// Handle root case
			if uid == "" {
				return true
			}

			// Check if this is a directory
			info, err := os.Stat(uid)
			if err != nil {
				return false
			}
			return info.IsDir()
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			icon.Resize(fyne.NewSize(16, 16)) // Set icon size
			label := widget.NewLabel("Template")
			// Remove truncation or set it to a more reasonable mode
			label.Truncation = fyne.TextTruncateOff
			return container.NewHBox(icon, label)
		},
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			// Update the node with appropriate icon and text
			hbox := node.(*fyne.Container)
			icon := hbox.Objects[0].(*widget.Icon)
			label := hbox.Objects[1].(*widget.Label)

			var displayName string
			if uid == "" {
				displayName = "Root"
			} else {
				displayName = filepath.Base(uid)
			}

			if branch {
				icon.SetResource(theme.FolderIcon())
			} else {
				// Set different icons based on file extension
				ext := filepath.Ext(displayName)
				switch ext {
				case ".nix":
					icon.SetResource(theme.DocumentIcon())
				case ".json":
					icon.SetResource(theme.DocumentIcon())
				case ".md":
					icon.SetResource(theme.DocumentIcon())
				default:
					icon.SetResource(theme.DocumentIcon())
				}
			}

			label.SetText(displayName)
		},
	)

	// Set up the selection handler
	tab.tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "" {
			// Root selected, clear editor
			tab.selectedFile = ""
			tab.editor.SetText("")
			tab.saveButton.Disable()
			return
		}

		tab.selectedFile = uid

		// Check if it's a file or directory
		info, err := os.Stat(uid)
		if err != nil {
			fyne.LogError("Failed to stat file: "+uid, err)
			tab.editor.SetText("")
			tab.saveButton.Disable()
			return
		}

		if info.IsDir() {
			// It's a directory, clear editor and disable save
			tab.editor.SetText("")
			tab.saveButton.Disable()
			return
		}

		// It's a file, read and display content
		content, err := os.ReadFile(uid)
		if err != nil {
			fyne.LogError("Failed to read file: "+uid, err)
			dialog.ShowError(err, win)
			tab.editor.SetText("")
			tab.saveButton.Disable()
			return
		}

		tab.editor.SetText(string(content))
		tab.saveButton.Enable()
	}

	// Create toolbar
	toolbar := container.NewHBox(
		refreshButton,
		widget.NewSeparator(),
		tab.saveButton,
	)

	// Create the tree with scroll
	treeContainer := container.NewBorder(
		toolbar, nil, nil, nil,
		container.NewScroll(tab.tree),
	)

	// Create editor container with line numbers (optional)
	editorScroll := container.NewScroll(tab.editor)
	editorContainer := container.NewBorder(
		nil, nil, nil, nil,
		editorScroll,
	)

	// Create split container
	split := container.NewHSplit(treeContainer, editorContainer)
	split.Offset = 0.35 // Make tree panel wider to accommodate filenames

	// Open flake.nix by default if it exists
	defaultFile := filepath.Join(rootPath, "flake.nix")
	if _, err := os.Stat(defaultFile); err == nil {
		// File exists, load it
		go func() {
			// Small delay to ensure UI is ready
			content, err := os.ReadFile(defaultFile)
			if err == nil {
				tab.selectedFile = defaultFile
				tab.editor.SetText(string(content))
				tab.saveButton.Enable()
				tab.tree.Select(defaultFile)
			}
		}()
	}

	tab.CanvasObject = split
	return tab
}
