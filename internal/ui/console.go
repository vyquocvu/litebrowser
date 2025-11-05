package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/vyquocvu/goosie/internal/js"
)

// ConsolePanel represents the developer console panel
type ConsolePanel struct {
	container     *fyne.Container
	messageList   *widget.List
	messages      []js.ConsoleMessage
	clearButton   *widget.Button
	filterSelect  *widget.Select
	filterLevel   string
	onRefresh     func()
	errorCountLabel *widget.Label
	errorCount    int
}

// NewConsolePanel creates a new console panel
func NewConsolePanel() *ConsolePanel {
	panel := &ConsolePanel{
		messages:    make([]js.ConsoleMessage, 0),
		filterLevel: "all",
		errorCount:  0,
	}

	// Create error count label
	panel.errorCountLabel = widget.NewLabel("Errors: 0")

	// Create clear button
	panel.clearButton = widget.NewButton("Clear", func() {
		panel.messages = make([]js.ConsoleMessage, 0)
		panel.errorCount = 0
		panel.errorCountLabel.SetText("Errors: 0")
		panel.messageList.Refresh()
		if panel.onRefresh != nil {
			panel.onRefresh()
		}
	})

	// Create filter dropdown
	panel.filterSelect = widget.NewSelect(
		[]string{"all", "log", "error", "warn", "info", "table"},
		func(selected string) {
			panel.filterLevel = selected
			panel.messageList.Refresh()
		},
	)
	panel.filterSelect.SetSelected("all")

	// Create message list
	panel.messageList = widget.NewList(
		func() int {
			return panel.getFilteredMessageCount()
		},
		func() fyne.CanvasObject {
			// Template for list items
			timeLabel := widget.NewLabel("")
			timeLabel.TextStyle.Monospace = true
			
			levelLabel := widget.NewLabel("")
			levelLabel.TextStyle.Bold = true
			
			messageLabel := widget.NewLabel("")
			messageLabel.Wrapping = fyne.TextWrapWord
			
			return container.NewBorder(
				container.NewHBox(timeLabel, levelLabel),
				nil, nil, nil,
				messageLabel,
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			msg := panel.getFilteredMessage(id)
			if msg == nil {
				return
			}

			border := item.(*fyne.Container)
			topContainer := border.Objects[0].(*fyne.Container)
			timeLabel := topContainer.Objects[0].(*widget.Label)
			levelLabel := topContainer.Objects[1].(*widget.Label)
			messageLabel := border.Objects[1].(*widget.Label)

			// Format time
			timeLabel.SetText(msg.Timestamp.Format("15:04:05"))

			// Set level with color
			switch msg.Level {
			case "error":
				levelLabel.SetText("[ERROR]")
				levelLabel.Importance = widget.HighImportance
			case "warn":
				levelLabel.SetText("[WARN]")
				levelLabel.Importance = widget.MediumImportance
			case "info":
				levelLabel.SetText("[INFO]")
				levelLabel.Importance = widget.LowImportance
			case "table":
				levelLabel.SetText("[TABLE]")
				levelLabel.Importance = widget.LowImportance
			default:
				levelLabel.SetText("[LOG]")
				levelLabel.Importance = widget.LowImportance
			}

			// Set message
			messageLabel.SetText(msg.Message)
		},
	)

	// Create toolbar
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewLabel("Filter:"),
			panel.filterSelect,
		),
		container.NewHBox(
			panel.errorCountLabel,
			panel.clearButton,
		),
		nil,
	)

	// Create main container with toolbar at top and scrollable message list
	panel.container = container.NewBorder(
		toolbar,
		nil, nil, nil,
		panel.messageList,
	)

	return panel
}

// GetContainer returns the console panel's container
func (cp *ConsolePanel) GetContainer() *fyne.Container {
	return cp.container
}

// AddMessage adds a new message to the console
func (cp *ConsolePanel) AddMessage(msg js.ConsoleMessage) {
	cp.messages = append(cp.messages, msg)
	
	// Update error count
	if msg.Level == "error" {
		cp.errorCount++
		cp.errorCountLabel.SetText(fmt.Sprintf("Errors: %d", cp.errorCount))
	}
	
	cp.messageList.Refresh()
}

// SetMessages replaces all messages in the console
func (cp *ConsolePanel) SetMessages(messages []js.ConsoleMessage) {
	cp.messages = messages
	
	// Count errors
	cp.errorCount = 0
	for _, msg := range messages {
		if msg.Level == "error" {
			cp.errorCount++
		}
	}
	cp.errorCountLabel.SetText(fmt.Sprintf("Errors: %d", cp.errorCount))
	
	cp.messageList.Refresh()
}

// Clear clears all console messages
func (cp *ConsolePanel) Clear() {
	cp.messages = make([]js.ConsoleMessage, 0)
	cp.errorCount = 0
	cp.errorCountLabel.SetText("Errors: 0")
	cp.messageList.Refresh()
}

// SetRefreshCallback sets the callback for when console is cleared
func (cp *ConsolePanel) SetRefreshCallback(callback func()) {
	cp.onRefresh = callback
}

// getFilteredMessageCount returns the count of messages matching the current filter
func (cp *ConsolePanel) getFilteredMessageCount() int {
	if cp.filterLevel == "all" {
		return len(cp.messages)
	}
	
	count := 0
	for _, msg := range cp.messages {
		if msg.Level == cp.filterLevel {
			count++
		}
	}
	return count
}

// getFilteredMessage returns the message at the given index after filtering
func (cp *ConsolePanel) getFilteredMessage(index int) *js.ConsoleMessage {
	if cp.filterLevel == "all" {
		if index >= 0 && index < len(cp.messages) {
			return &cp.messages[index]
		}
		return nil
	}
	
	// Find the nth message matching the filter
	currentIndex := 0
	for i := range cp.messages {
		if cp.messages[i].Level == cp.filterLevel {
			if currentIndex == index {
				return &cp.messages[i]
			}
			currentIndex++
		}
	}
	return nil
}

// GetErrorCount returns the current error count
func (cp *ConsolePanel) GetErrorCount() int {
	return cp.errorCount
}
