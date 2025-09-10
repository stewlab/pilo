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
	CanvasObject  fyne.CanvasObject
	tree          *widget.Tree
	editor        *widget.Entry
	selectedFile  string
	saveButton    *widget.Button
	treeContainer *fyne.Container
	win           fyne.Window
}

func (t *ConfigEditorTab) Refresh() {
	// Store the currently selected file
	currentFile := t.selectedFile

	// Create a new tree
	newTree := t.createFileTree()

	// Replace the old tree with the new one
	t.tree = newTree
	t.treeContainer.Objects[0] = container.NewScroll(t.tree)
	t.treeContainer.Refresh()

	// If a file was previously selected, re-select it
	if currentFile != "" {
		t.tree.Select(currentFile)
		// Reload the file content
		content, err := os.ReadFile(currentFile)
		if err == nil {
			t.editor.SetText(string(content))
		}
	}
}

func (t *ConfigEditorTab) createFileTree() *widget.Tree {
	rootPath := config.GetFlakePath()
	tree := widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			var path string
			if uid == "" {
				path = rootPath
			} else {
				path = uid
			}
			if _, err := os.Stat(path); err != nil {
				fyne.LogError("Path does not exist: "+path, err)
				return []widget.TreeNodeID{}
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				fyne.LogError("Failed to read directory: "+path, err)
				return []widget.TreeNodeID{}
			}
			var children []widget.TreeNodeID
			for _, entry := range entries {
				if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
					continue
				}
				childPath := filepath.Join(path, entry.Name())
				children = append(children, childPath)
			}
			return children
		},
		func(uid widget.TreeNodeID) bool {
			if uid == "" {
				return true
			}
			info, err := os.Stat(uid)
			if err != nil {
				return false
			}
			return info.IsDir()
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			icon.Resize(fyne.NewSize(16, 16))
			label := widget.NewLabel("Template")
			label.Truncation = fyne.TextTruncateOff
			return container.NewHBox(icon, label)
		},
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
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
				ext := filepath.Ext(displayName)
				switch ext {
				case ".nix", ".json", ".md":
					icon.SetResource(theme.DocumentIcon())
				default:
					icon.SetResource(theme.DocumentIcon())
				}
			}
			label.SetText(displayName)
		},
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "" {
			t.selectedFile = ""
			t.editor.SetText("")
			t.saveButton.Disable()
			return
		}
		t.selectedFile = uid
		info, err := os.Stat(uid)
		if err != nil {
			fyne.LogError("Failed to stat file: "+uid, err)
			t.editor.SetText("")
			t.saveButton.Disable()
			return
		}
		if info.IsDir() {
			t.editor.SetText("")
			t.saveButton.Disable()
			return
		}
		content, err := os.ReadFile(uid)
		if err != nil {
			fyne.LogError("Failed to read file: "+uid, err)
			dialog.ShowError(err, t.win)
			t.editor.SetText("")
			t.saveButton.Disable()
			return
		}
		t.editor.SetText(string(content))
		t.saveButton.Enable()
	}
	return tree
}

func CreateConfigEditorTab(win fyne.Window) *ConfigEditorTab {
	tab := &ConfigEditorTab{win: win}
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

	refreshButton := widget.NewButton("Refresh", func() {
		tab.Refresh()
	})

	tab.tree = tab.createFileTree()

	toolbar := container.NewHBox(
		refreshButton,
		widget.NewSeparator(),
		tab.saveButton,
	)

	tab.treeContainer = container.NewBorder(
		toolbar, nil, nil, nil,
		container.NewScroll(tab.tree),
	)

	editorScroll := container.NewScroll(tab.editor)
	editorContainer := container.NewBorder(
		nil, nil, nil, nil,
		editorScroll,
	)

	split := container.NewHSplit(tab.treeContainer, editorContainer)
	split.Offset = 0.35

	defaultFile := filepath.Join(config.GetFlakePath(), "flake.nix")
	if _, err := os.Stat(defaultFile); err == nil {
		go func() {
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
