package ui

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Browser represents the browser UI
type Browser struct {
	app        app.App
	window     window
	contentBox *widget.Label
}

// window interface to allow testing
type window interface {
	SetTitle(string)
	SetContent(content interface{})
	ShowAndRun()
	Resize(size interface{})
}

// NewBrowser creates a new browser UI
func NewBrowser() *Browser {
	a := app.New()
	w := a.NewWindow("Goja Browser")
	
	contentBox := widget.NewLabel("Loading...")
	contentBox.Wrapping = 1 // fyne.TextWrapWord
	
	return &Browser{
		app:        a,
		window:     w,
		contentBox: contentBox,
	}
}

// SetContent updates the displayed content
func (b *Browser) SetContent(content string) {
	b.contentBox.SetText(content)
}

// Show displays the browser window
func (b *Browser) Show() {
	scroll := container.NewScroll(b.contentBox)
	b.window.SetContent(scroll)
	b.window.ShowAndRun()
}
