package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Browser represents the browser UI
type Browser struct {
	app        fyne.App
	window     fyne.Window
	contentBox *widget.RichText
}

// window interface to allow testing
type window interface {
	SetTitle(string)
	SetContent(fyne.CanvasObject)
	ShowAndRun()
	Resize(fyne.Size)
}

// NewBrowser creates a new browser UI
func NewBrowser() *Browser {
	a := app.New()
	w := a.NewWindow("Goja Browser")
	
	// Set window size
	w.Resize(fyne.NewSize(800, 600))
	
	contentBox := widget.NewRichTextFromMarkdown("Loading...")
	contentBox.Wrapping = fyne.TextWrapWord
	
	return &Browser{
		app:        a,
		window:     w,
		contentBox: contentBox,
	}
}

// SetContent updates the displayed content (plain text)
func (b *Browser) SetContent(content string) {
	b.contentBox.ParseMarkdown(content)
}

// SetHTMLContent updates the displayed content from markdown-formatted HTML
func (b *Browser) SetHTMLContent(content string) {
	b.contentBox.ParseMarkdown(content)
}

// Show displays the browser window
func (b *Browser) Show() {
	scroll := container.NewScroll(b.contentBox)
	b.window.SetContent(scroll)
	b.window.ShowAndRun()
}
