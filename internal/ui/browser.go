package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	state               *BrowserState
	settings            *Settings
	urlEntry            *widget.Entry
	backButton          *widget.Button
	forwardButton       *widget.Button
	refreshButton       *widget.Button
	bookmarkButton      *widget.Button
	settingsButton      *widget.Button
	loadingBar          *widget.ProgressBarInfinite
	loadingBarContainer *fyne.Container
	onNavigate          NavigationCallback
	tabs                *container.DocTabs
	tabItems            []*Tab
}

// Tab represents a single browser tab
type Tab struct {
	title         string
	content       fyne.CanvasObject
	contentBox    *widget.RichText
	contentScroll *container.Scroll
	htmlRenderer  *renderer.Renderer
	state         *BrowserState
	browser       *Browser
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

	state := NewBrowserState()
	settings := NewSettings()

	// Create thin, full-width loading progress bar with 5px height (initially hidden)
	loadingBar := widget.NewProgressBarInfinite()
	loadingBar.Hide()

	// Wrap the progress bar in a container with fixed height of 5px
	loadingBarContainer := container.New(&fixedHeightLayout{height: 5}, loadingBar)
	loadingBarContainer.Hide()

	browser := &Browser{
		app:                 a,
		window:              w,
		state:               state,
		settings:            settings,
		loadingBar:          loadingBar,
		loadingBarContainer: loadingBarContainer,
		tabItems:            []*Tab{},
	}

	firstTab := browser.newTabInternal()
	browser.tabItems = append(browser.tabItems, firstTab)

	browser.tabs = container.NewDocTabs(firstTab.AsTabItem())
	browser.tabs.CreateTab = func() *container.TabItem {
		tab := browser.NewTab()
		return tab.AsTabItem()
	}
	browser.tabs.OnSelected = func(tab *container.TabItem) {
		browser.updateNavigationButtons()
	}
	browser.tabs.SetTabLocation(container.TabLocationTop)

	browser.createNavigationControls()

	return browser
}

// newTabInternal creates a new tab without adding it to the tab container
func (b *Browser) newTabInternal() *Tab {
	contentBox := widget.NewRichTextFromMarkdown("Welcome to Goosie! Enter a URL above to start browsing.")
	contentBox.Wrapping = fyne.TextWrapWord
	contentScroll := container.NewScroll(contentBox)

	htmlRenderer := renderer.NewRenderer(1000, 700)
	htmlRenderer.SetWindow(b.window)
	htmlRenderer.SetNavigationCallback(func(url string) {
		if b.onNavigate != nil {
			b.onNavigate(url)
		}
	})

	tabState := NewBrowserState()

	return &Tab{
		title:         "New Tab",
		content:       contentScroll,
		contentBox:    contentBox,
		contentScroll: contentScroll,
		htmlRenderer:  htmlRenderer,
		state:         tabState,
		browser:       b,
	}
}

// NewTab creates a new browser tab and adds it to the tab container
func (b *Browser) NewTab() *Tab {
	tab := b.newTabInternal()
	b.tabItems = append(b.tabItems, tab)
	return tab
}

// ActiveTab returns the currently active tab
func (b *Browser) ActiveTab() *Tab {
	if len(b.tabItems) == 0 {
		return nil
	}
	selectedIndex := b.tabs.SelectedIndex()
	if selectedIndex < 0 || selectedIndex >= len(b.tabItems) {
		return nil
	}
	return b.tabItems[selectedIndex]
}

// SetContent updates the displayed content (plain text)
func (b *Browser) SetContent(content string) {
	if tab := b.ActiveTab(); tab != nil {
		tab.contentBox.ParseMarkdown(content)
	}
}

// SetHTMLContent updates the displayed content from markdown-formatted HTML
func (b *Browser) SetHTMLContent(content string) {
	if tab := b.ActiveTab(); tab != nil {
		tab.contentBox.ParseMarkdown(content)
	}
}

// RenderHTMLContent renders HTML content using the canvas-based renderer
func (b *Browser) RenderHTMLContent(htmlContent string) error {
	tab := b.ActiveTab()
	if tab == nil {
		return nil
	}
	// Set the current URL for resolving relative links
	currentURL := tab.state.GetCurrentURL()
	tab.htmlRenderer.SetCurrentURL(currentURL)

	canvasObject, err := tab.htmlRenderer.RenderHTML(htmlContent)
	if err != nil {
		return err
	}

	// Update the scroll container with the rendered content on the main thread
	fyne.Do(func() {
		tab.contentScroll.Content = canvasObject
		tab.contentScroll.Refresh()
	})

	return nil
}

// SetNavigationCallback sets the callback for when navigation is requested
func (b *Browser) SetNavigationCallback(callback NavigationCallback) {
	b.onNavigate = callback
}

// Show displays the browser window
func (b *Browser) Show() {
	// Create navigation bar
	navBar := container.NewBorder(nil, nil,
		container.NewHBox(b.backButton, b.forwardButton, b.refreshButton),
		container.NewHBox(b.bookmarkButton, b.settingsButton),
		b.urlEntry,
	)

	// Create main layout with 5px height loading bar
	content := container.NewBorder(
		container.NewVBox(navBar, b.loadingBarContainer),
		nil, nil, nil,
		b.tabs,
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
		if tab := b.ActiveTab(); tab != nil {
			if url, ok := tab.state.GoBack(); ok {
				if b.onNavigate != nil {
					b.onNavigate(url)
				}
			}
		}
	})
	b.backButton.Disable()

	// Forward button
	b.forwardButton = widget.NewButton("→", func() {
		if tab := b.ActiveTab(); tab != nil {
			if url, ok := tab.state.GoForward(); ok {
				if b.onNavigate != nil {
					b.onNavigate(url)
				}
			}
		}
	})
	b.forwardButton.Disable()

	// Refresh button
	b.refreshButton = widget.NewButton("⟳", func() {
		if tab := b.ActiveTab(); tab != nil {
			currentURL := tab.state.GetCurrentURL()
			if b.onNavigate != nil && currentURL != "" {
				b.onNavigate(currentURL)
			}
		}
	})

	// Bookmark button
	b.bookmarkButton = widget.NewButton("☆", func() {
		b.toggleBookmark()
	})
	b.bookmarkButton.Disable()

	// Settings button
	b.settingsButton = widget.NewButton("⚙", func() {
		b.showSettings()
	})
}

// AsTabItem converts a Tab to a TabItem
func (t *Tab) AsTabItem() *container.TabItem {
	return container.NewTabItem(t.title, t.content)
}

// toggleBookmark adds or removes the current page from bookmarks
func (b *Browser) toggleBookmark() {
	if tab := b.ActiveTab(); tab != nil {
		currentURL := tab.state.GetCurrentURL()
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
}

// NavigateTo navigates to a URL and updates the UI
func (b *Browser) NavigateTo(url string) {
	if tab := b.ActiveTab(); tab != nil {
		tab.state.AddToHistory(url)
		b.urlEntry.SetText(url)
		b.updateNavigationButtons()
	}
}

// updateNavigationButtons updates the enabled/disabled state of navigation buttons
func (b *Browser) updateNavigationButtons() {
	tab := b.ActiveTab()
	if tab == nil {
		b.backButton.Disable()
		b.forwardButton.Disable()
		b.bookmarkButton.Disable()
		return
	}

	if tab.state.CanGoBack() {
		b.backButton.Enable()
	} else {
		b.backButton.Disable()
	}

	if tab.state.CanGoForward() {
		b.forwardButton.Enable()
	} else {
		b.forwardButton.Disable()
	}

	currentURL := tab.state.GetCurrentURL()
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
	if tab := b.ActiveTab(); tab != nil {
		return tab.state.GetHistory()
	}
	return []string{}
}

// ShowLoading displays the loading indicator
func (b *Browser) ShowLoading() {
	// Use fyne.Do to ensure UI updates happen on the main thread
	fyne.Do(func() {
		b.loadingBarContainer.Show()
		b.loadingBar.Show()
	})
}

// HideLoading hides the loading indicator
func (b *Browser) HideLoading() {
	// Use fyne.Do to ensure UI updates happen on the main thread
	fyne.Do(func() {
		b.loadingBar.Hide()
		b.loadingBarContainer.Hide()
	})
}

// UpdateActiveTabTitle updates the title of the active tab
func (b *Browser) UpdateActiveTabTitle(title string) {
	fyne.Do(func() {
		if tab := b.ActiveTab(); tab != nil {
			tab.title = title
			if selected := b.tabs.Selected(); selected != nil {
				selected.Text = title
				b.tabs.Refresh()
			}
		}
	})
}

// GetApp returns the Fyne application instance for thread-safe operations
func (b *Browser) GetApp() fyne.App {
	return b.app
}

// GetSettings returns the browser settings
func (b *Browser) GetSettings() *Settings {
	return b.settings
}

// showSettings displays the settings dialog
func (b *Browser) showSettings() {
	// Create form entries for settings
	homepageEntry := widget.NewEntry()
	homepageEntry.SetText(b.settings.GetHomepage())
	homepageEntry.SetPlaceHolder("https://example.com")

	searchEngineEntry := widget.NewEntry()
	searchEngineEntry.SetText(b.settings.GetDefaultSearchEngine())
	searchEngineEntry.SetPlaceHolder("https://www.google.com/search?q=")

	jsCheck := widget.NewCheck("Enable JavaScript", func(checked bool) {
		b.settings.SetEnableJavaScript(checked)
	})
	jsCheck.SetChecked(b.settings.GetEnableJavaScript())

	imagesCheck := widget.NewCheck("Enable Images", func(checked bool) {
		b.settings.SetEnableImages(checked)
	})
	imagesCheck.SetChecked(b.settings.GetEnableImages())

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Homepage", Widget: homepageEntry},
			{Text: "Search Engine", Widget: searchEngineEntry},
			{Text: "", Widget: jsCheck},
			{Text: "", Widget: imagesCheck},
		},
		OnSubmit: func() {
			// Save settings
			b.settings.SetHomepage(homepageEntry.Text)
			b.settings.SetDefaultSearchEngine(searchEngineEntry.Text)
		},
		OnCancel: func() {
			// Do nothing, just close
		},
	}

	// Create custom dialog
	d := dialog.NewCustom("Settings", "Close", form, b.window)
	d.Resize(fyne.NewSize(500, 300))
	d.Show()
}
