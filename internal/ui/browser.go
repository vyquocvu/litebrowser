package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/vyquocvu/goosie/internal/renderer"
)

// fixedHeightLayout is a custom layout that sets a fixed height for a widget
type fixedHeightLayout struct {
	height float32
}

func (l *fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, l.height)
	}
	// Use the widget's minimum width but override the height
	minSize := objects[0].MinSize()
	return fyne.NewSize(minSize.Width, l.height)
}

func (l *fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	// Position the widget to fill the width but constrain height
	objects[0].Resize(fyne.NewSize(size.Width, l.height))
	objects[0].Move(fyne.NewPos(0, 0))
}

// NavigationCallback is a function that is called when navigation is requested
type NavigationCallback func(url string)

// Browser represents the browser UI
type Browser struct {
	app                 fyne.App
	window              fyne.Window
	contentBox          *widget.RichText
	contentScroll       *container.Scroll
	state               *BrowserState
	urlEntry            *widget.Entry
	backButton          *widget.Button
	forwardButton       *widget.Button
	refreshButton       *widget.Button
	bookmarkButton      *widget.Button
	loadingBar          *widget.ProgressBarInfinite
	loadingBarContainer *fyne.Container
	onNavigate          NavigationCallback
	htmlRenderer        *renderer.Renderer
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
	w := a.NewWindow("Goosie")

	// Set window size
	w.Resize(fyne.NewSize(1000, 700))

	contentBox := widget.NewRichTextFromMarkdown("Welcome to Goosie! Enter a URL above to start browsing.")
	contentBox.Wrapping = fyne.TextWrapWord

	state := NewBrowserState()

	// Create HTML renderer with canvas size
	htmlRenderer := renderer.NewRenderer(1000, 700)

	// Create scroll container
	contentScroll := container.NewScroll(contentBox)

	// Create thin, full-width loading progress bar with 5px height (initially hidden)
	loadingBar := widget.NewProgressBarInfinite()
	loadingBar.Hide()

	// Wrap the progress bar in a container with fixed height of 5px
	loadingBarContainer := container.New(&fixedHeightLayout{height: 5}, loadingBar)
	loadingBarContainer.Hide()

	browser := &Browser{
		app:                 a,
		window:              w,
		contentBox:          contentBox,
		contentScroll:       contentScroll,
		state:               state,
		htmlRenderer:        htmlRenderer,
		loadingBar:          loadingBar,
		loadingBarContainer: loadingBarContainer,
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
	// Set the current URL for resolving relative links
	currentURL := b.state.GetCurrentURL()
	b.htmlRenderer.SetCurrentURL(currentURL)

	canvasObject, err := b.htmlRenderer.RenderHTML(htmlContent)
	if err != nil {
		return err
	}

	// Update the scroll container with the rendered content on the main thread
	fyne.Do(func() {
		b.contentScroll.Content = canvasObject

		// Get content height and update viewport
		contentHeight := b.htmlRenderer.GetContentHeight()
		if contentHeight > 0 {
			// Initialize viewport to full height
			b.htmlRenderer.SetViewport(0, b.contentScroll.Size().Height)
		}

		b.contentScroll.Refresh()
	})

	return nil
}

// SetNavigationCallback sets the callback for when navigation is requested
func (b *Browser) SetNavigationCallback(callback NavigationCallback) {
	b.onNavigate = callback
	// Also pass the callback to the renderer for link clicks
	b.htmlRenderer.SetNavigationCallback(func(url string) {
		if b.onNavigate != nil {
			b.onNavigate(url)
		}
	})
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

	// Create main layout with 5px height loading bar
	content := container.NewBorder(
		container.NewVBox(navBar, b.loadingBarContainer),
		nil, nil, nil,
		b.contentScroll,
	)

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

// ShowLoading displays the loading indicator
func (b *Browser) ShowLoading() {
	// Use fyne.Do to ensure UI updates happen on the main thread
	fyne.Do(func() {
		b.loadingBarContainer.Show()
		b.loadingBar.Show()
		b.loadingBar.Start()
	})
}

// HideLoading hides the loading indicator
func (b *Browser) HideLoading() {
	// Use fyne.Do to ensure UI updates happen on the main thread
	fyne.Do(func() {
		b.loadingBar.Stop()
		b.loadingBar.Hide()
		b.loadingBarContainer.Hide()
	})
}

// GetApp returns the Fyne application instance for thread-safe operations
func (b *Browser) GetApp() fyne.App {
	return b.app
}
