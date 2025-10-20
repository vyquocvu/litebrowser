package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/vyquocvu/litebrowser/internal/renderer"
)

// NavigationCallback is a function that is called when navigation is requested
type NavigationCallback func(url string)

// Browser represents the browser UI
type Browser struct {
	app            fyne.App
	window         fyne.Window
	contentBox     *widget.RichText
	contentScroll  *container.Scroll
	state          *BrowserState
	urlEntry       *widget.Entry
	backButton     *widget.Button
	forwardButton  *widget.Button
	refreshButton  *widget.Button
	bookmarkButton *widget.Button
	onNavigate     NavigationCallback
	htmlRenderer   *renderer.Renderer
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
	w := a.NewWindow("Litebrowser")

	// Set window size
	w.Resize(fyne.NewSize(1000, 700))

	contentBox := widget.NewRichTextFromMarkdown("Welcome to Litebrowser! Enter a URL above to start browsing.")
	contentBox.Wrapping = fyne.TextWrapWord

	state := NewBrowserState()

	// Create HTML renderer with canvas size
	htmlRenderer := renderer.NewRenderer(1000, 700)

	// Create scroll container
	contentScroll := container.NewScroll(contentBox)

	browser := &Browser{
		app:           a,
		window:        w,
		contentBox:    contentBox,
		contentScroll: contentScroll,
		state:         state,
		htmlRenderer:  htmlRenderer,
	}

	return browser
}

// SetContent updates the displayed content (plain text)
func (b *Browser) SetContent(content string) {
	b.contentBox.ParseMarkdown(content)
}

// SetHTMLContent updates the displayed content from markdown-formatted HTML
func (b *Browser) SetHTMLContent(content string) {
	b.contentBox.ParseMarkdown(content)
}

// RenderHTMLContent renders HTML content using the canvas-based renderer
func (b *Browser) RenderHTMLContent(htmlContent string) error {
	canvasObject, err := b.htmlRenderer.RenderHTML(htmlContent)
	if err != nil {
		return err
	}
	
	// Update the scroll container with the rendered content
	b.contentScroll.Content = canvasObject
	b.contentScroll.Refresh()
	
	return nil
}

// SetNavigationCallback sets the callback for when navigation is requested
func (b *Browser) SetNavigationCallback(callback NavigationCallback) {
	b.onNavigate = callback
}

// Show displays the browser window
func (b *Browser) Show() {
	// Create navigation controls
	b.createNavigationControls()

	// Create navigation bar
	navBar := container.NewBorder(nil, nil,
		container.NewHBox(b.backButton, b.forwardButton, b.refreshButton),
		container.NewHBox(b.bookmarkButton),
		b.urlEntry,
	)

	// Create main layout
	content := container.NewBorder(navBar, nil, nil, nil, b.contentScroll)

	b.window.SetContent(content)
	b.window.ShowAndRun()
}

// createNavigationControls creates all navigation UI controls
func (b *Browser) createNavigationControls() {
	// URL entry
	b.urlEntry = widget.NewEntry()
	b.urlEntry.SetPlaceHolder("Enter URL (e.g., https://example.com)")
	b.urlEntry.OnSubmitted = func(url string) {
		if b.onNavigate != nil && url != "" {
			b.onNavigate(url)
		}
	}

	// Back button
	b.backButton = widget.NewButton("←", func() {
		if url, ok := b.state.GoBack(); ok {
			if b.onNavigate != nil {
				b.onNavigate(url)
			}
		}
	})
	b.backButton.Disable()

	// Forward button
	b.forwardButton = widget.NewButton("→", func() {
		if url, ok := b.state.GoForward(); ok {
			if b.onNavigate != nil {
				b.onNavigate(url)
			}
		}
	})
	b.forwardButton.Disable()

	// Refresh button
	b.refreshButton = widget.NewButton("⟳", func() {
		currentURL := b.state.GetCurrentURL()
		if b.onNavigate != nil && currentURL != "" {
			b.onNavigate(currentURL)
		}
	})

	// Bookmark button
	b.bookmarkButton = widget.NewButton("☆", func() {
		b.toggleBookmark()
	})
	b.bookmarkButton.Disable()
}

// toggleBookmark adds or removes the current page from bookmarks
func (b *Browser) toggleBookmark() {
	currentURL := b.state.GetCurrentURL()
	if currentURL == "" {
		return
	}

	if b.state.IsBookmarked(currentURL) {
		b.state.RemoveBookmark(currentURL)
		b.bookmarkButton.SetText("☆")
	} else {
		b.state.AddBookmark(currentURL)
		b.bookmarkButton.SetText("★")
	}
	b.bookmarkButton.Refresh()
}

// NavigateTo navigates to a URL and updates the UI
func (b *Browser) NavigateTo(url string) {
	b.state.AddToHistory(url)
	b.urlEntry.SetText(url)
	b.updateNavigationButtons()
}

// updateNavigationButtons updates the enabled/disabled state of navigation buttons
func (b *Browser) updateNavigationButtons() {
	if b.state.CanGoBack() {
		b.backButton.Enable()
	} else {
		b.backButton.Disable()
	}

	if b.state.CanGoForward() {
		b.forwardButton.Enable()
	} else {
		b.forwardButton.Disable()
	}

	currentURL := b.state.GetCurrentURL()
	if currentURL != "" {
		b.bookmarkButton.Enable()
		if b.state.IsBookmarked(currentURL) {
			b.bookmarkButton.SetText("★")
		} else {
			b.bookmarkButton.SetText("☆")
		}
		b.bookmarkButton.Refresh()
	} else {
		b.bookmarkButton.Disable()
	}
}

// GetBookmarks returns the list of bookmarks
func (b *Browser) GetBookmarks() []string {
	return b.state.GetBookmarks()
}

// GetHistory returns the navigation history
func (b *Browser) GetHistory() []string {
	return b.state.GetHistory()
}
