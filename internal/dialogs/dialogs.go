package dialogs

import (
	"fmt"
	"pilo/internal/api"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Enhanced log viewer using RichText
func ShowGitStatusDialog(status string, win fyne.Window, repoPath string) {
	// Status tab with RichText
	statusText := widget.NewRichTextFromMarkdown(fmt.Sprintf("```\n%s\n```", status))
	statusText.Wrapping = fyne.TextWrapWord

	statusCopyButton := widget.NewButton("Copy All", func() {
		fyne.CurrentApp().Clipboard().SetContent(status)
	})

	clearButton := widget.NewButton("Clear", func() {
		statusText.ParseMarkdown("```\n\n```")
	})

	buttonContainer := container.NewHBox(statusCopyButton, clearButton)
	statusContent := container.NewBorder(nil, buttonContainer, nil, nil,
		container.NewScroll(statusText))

	// Diff tab with RichText (loaded lazily)
	diffText := widget.NewRichText()
	diffText.Wrapping = fyne.TextWrapWord

	var diffContent string
	diffCopyButton := widget.NewButton("Copy Diff", func() {
		if diffContent != "" {
			fyne.CurrentApp().Clipboard().SetContent(diffContent)
		}
	})

	diffButtonContainer := container.NewHBox(diffCopyButton)
	diffContainer := container.NewBorder(nil, diffButtonContainer, nil, nil,
		container.NewScroll(diffText))

	tabs := container.NewAppTabs(
		container.NewTabItem("Status", statusContent),
		container.NewTabItem("Diff", diffContainer),
	)

	var diffLoaded bool
	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Diff" && !diffLoaded {
			// Show loading indicator
			diffText.ParseMarkdown("*Loading diff...*")

			go func() {
				diff, err := api.GetGitDiff(repoPath)
				if err != nil {
					diffContent = fmt.Sprintf("Error getting diff: %v", err)
					fyne.Do(func() {
						diffText.ParseMarkdown(fmt.Sprintf("**Error:** %s", err.Error()))
					})
				} else {
					diffContent = diff
					fyne.Do(func() {
						// Use code block for syntax highlighting
						diffText.ParseMarkdown(fmt.Sprintf("```diff\n%s\n```", diff))
					})
				}
				diffLoaded = true
			}()
		}
	}

	tabs.SetTabLocation(container.TabLocationTop)
	content := container.NewBorder(nil, nil, nil, nil, tabs)

	d := dialog.NewCustom("Git Status", "Close", content, win)
	d.Resize(fyne.NewSize(900, 700))
	d.Show()
}

// LogViewer is a dedicated component for continuous logging.
type LogViewer struct {
	widget.BaseWidget
	richText   *widget.RichText
	content    strings.Builder
	maxLines   int
	autoScroll bool
	scroll     *container.Scroll
}

// NewLogViewer creates a new LogViewer instance.
func NewLogViewer() *LogViewer {
	l := &LogViewer{
		richText:   widget.NewRichText(),
		maxLines:   1000, // Limit to prevent memory issues
		autoScroll: true,
	}
	l.richText.Wrapping = fyne.TextWrapWord
	l.ExtendBaseWidget(l)
	return l
}

// CreateRenderer creates the widget renderer.
func (l *LogViewer) CreateRenderer() fyne.WidgetRenderer {
	l.scroll = container.NewScroll(l.richText)
	l.scroll.SetMinSize(fyne.NewSize(400, 300))

	copyBtn := widget.NewButton("Copy All", func() {
		fyne.CurrentApp().Clipboard().SetContent(l.content.String())
	})

	clearBtn := widget.NewButton("Clear", func() {
		l.Clear()
	})

	autoScrollCheck := widget.NewCheck("Auto-scroll", func(checked bool) {
		l.autoScroll = checked
	})
	autoScrollCheck.SetChecked(l.autoScroll)

	toolbar := container.NewHBox(copyBtn, clearBtn, autoScrollCheck)

	content := container.NewBorder(nil, toolbar, nil, nil, l.scroll)

	return widget.NewSimpleRenderer(content)
}

// AppendLog adds a new log line to the viewer.
func (l *LogViewer) AppendLog(text string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, text)

	l.content.WriteString(logLine)

	lines := strings.Split(l.content.String(), "\n")
	if len(lines) > l.maxLines {
		trimmed := lines[len(lines)-l.maxLines:]
		l.content.Reset()
		l.content.WriteString(strings.Join(trimmed, "\n"))
	}

	l.richText.ParseMarkdown(fmt.Sprintf("```\n%s```", l.content.String()))

	if l.autoScroll && l.scroll != nil {
		l.scroll.ScrollToBottom()
	}
}

// Clear removes all log entries.
func (l *LogViewer) Clear() {
	l.content.Reset()
	l.richText.ParseMarkdown("```\n\n```")
}

// SetMaxLines sets the maximum number of lines to retain.
func (l *LogViewer) SetMaxLines(max int) {
	l.maxLines = max
}

// ShowLogsDialog displays logs using the LogViewer.
func ShowLogsDialog(logs binding.StringList, win fyne.Window) {
	logViewer := NewLogViewer()

	logData, _ := logs.Get()
	for _, line := range logData {
		logViewer.AppendLog(line)
	}

	logs.AddListener(binding.NewDataListener(func() {
		newData, _ := logs.Get()
		if len(newData) > 0 {
			// Assuming new logs are appended at the end
			lastLog := newData[len(newData)-1]
			logViewer.AppendLog(lastLog)
		}
	}))

	d := dialog.NewCustom("📜  Logs", "Close", logViewer, win)
	d.Resize(fyne.NewSize(800, 600))
	d.Show()
}

// ShowPasswordDialog shows a dialog to ask for a password.
func ShowPasswordDialog(win fyne.Window, onConfirm func(password string)) {
	passwordEntry := widget.NewPasswordEntry()
	form := dialog.NewForm("Sudo Password Required", "Confirm", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Sudo Password", passwordEntry),
	}, func(ok bool) {
		if !ok {
			return
		}
		onConfirm(passwordEntry.Text)
	}, win)
	form.Resize(fyne.NewSize(400, 150))
	form.Show()
}

// ShowRunningCommandDialog shows a dialog for a running command using LogViewer.
func ShowRunningCommandDialog(
	win fyne.Window,
	title string,
	action func() (string, error),
	onClose func(output string, err error),
) {
	logViewer := NewLogViewer()
	progress := widget.NewProgressBarInfinite()
	content := container.NewBorder(progress, nil, nil, nil, logViewer)

	var result string
	var err error

	d := dialog.NewCustom(title, "Cancel", content, win)
	d.SetOnClosed(func() {
		if onClose != nil {
			onClose(result, err)
		}
	})

	go func() {
		result, err = action()
		var outputLines []string
		if err != nil {
			outputLines = strings.Split(fmt.Sprintf("Error: %v\n\n%s", err, result), "\n")
		} else {
			outputLines = append([]string{"Command executed successfully!"}, strings.Split(result, "\n")...)
		}

		fyne.Do(func() {
			for _, line := range outputLines {
				logViewer.AppendLog(line)
			}
			progress.Hide()
			d.SetDismissText("Close")
		})
	}()

	d.Resize(fyne.NewSize(800, 600))
	d.Show()
}

// ShowForm shows a form dialog.
func ShowForm(win fyne.Window, title, confirm, dismiss string, items []*widget.FormItem, callback func(bool)) {
	dialog.ShowForm(title, confirm, dismiss, items, callback, win)
}

// ShowConfirm shows a confirmation dialog.
func ShowConfirm(win fyne.Window, title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, win)
}

// ShowErrorDialog shows a dialog to display an error message.
func ShowErrorDialog(err error, win fyne.Window) {
	dialog.ShowError(err, win)
}

// ShowCustomConfirm shows a custom confirmation dialog.
func ShowCustomConfirm(win fyne.Window, title, confirm, dismiss string, content fyne.CanvasObject, callback func(bool)) {
	dialog.NewCustomConfirm(title, confirm, dismiss, content, callback, win).Show()
}

// ShowCommitDialog shows a dialog for entering a commit message.
func ShowCommitDialog(win fyne.Window, onConfirm func(message string)) {
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Enter commit message...")
	entry.SetMinRowsVisible(3)

	form := dialog.NewForm("Commit Changes", "Commit", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Commit Message", entry),
	}, func(ok bool) {
		if !ok {
			return
		}
		onConfirm(entry.Text)
	}, win)
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}
