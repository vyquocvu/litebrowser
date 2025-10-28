package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/vyquocvu/goosie/internal/dom"
	"github.com/vyquocvu/goosie/internal/js"
	"github.com/vyquocvu/goosie/internal/net"
	"github.com/vyquocvu/goosie/internal/ui"
	"golang.org/x/net/html"
)

func main() {
	// Initialize components
	fetcher := net.NewFetcher()
	parser := dom.NewParser()
	jsRuntime := js.NewRuntime()
	browser := ui.NewBrowser()

	// Create a cancellable context for page loads
	var currentLoadCtx context.Context
	var currentLoadCancel context.CancelFunc

	// Set up navigation callback
	browser.SetNavigationCallback(func(url string) {
		// Cancel any ongoing page load
		if currentLoadCancel != nil {
			currentLoadCancel()
		}

		// Create new context for this load
		currentLoadCtx, currentLoadCancel = context.WithCancel(context.Background())

		// Load page asynchronously
		loadPageAsync(browser, fetcher, parser, jsRuntime, url, currentLoadCtx)
	})

	// Show browser window
	browser.Show()
}

// pageLoadResult represents the result of an async page load
type pageLoadResult struct {
	html string
	err  error
}

// loadPageAsync fetches and displays a web page asynchronously
func loadPageAsync(browser *ui.Browser, fetcher *net.Fetcher, parser *dom.Parser, jsRuntime *js.Runtime, url string, ctx context.Context) {
	log.Printf("Navigating to: %s", url)

	// Update browser state on main thread
	browser.NavigateTo(url)

	// Show loading indicator on main thread
	browser.ShowLoading()

	// Launch background goroutine for fetch and render
	go func() {
		// Fetch the page in background
		html, err := fetcher.FetchWithContext(ctx, url, func(progress float64) {
			browser.UpdateLoadingProgress(progress)
		})

		// Check if context was cancelled
		if ctx.Err() != nil {
			log.Printf("Page load cancelled for: %s", url)
			return
		}

		if err != nil {
			// Fallback to mock HTML for example.com if network is unavailable
			log.Printf("Network error (%v), checking if example.com for mock HTML", err)
			if url == "https://example.com" {
				html = `<!DOCTYPE html>
<html>
<head>
    <title>Example Domain</title>
</head>
<body>
    <div>
        <h1>Example Domain</h1>
        <p id="main-content">This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.</p>
        <p><a href="https://www.iana.org/domains/example">More information...</a></p>
    </div>
</body>
</html>`
			} else {
				// Update UI on main thread with error
				updateUIWithError(browser, err, url)
				return
			}
		}

		// Update UI on main thread with content
		updateUIWithContent(browser, jsRuntime, html, url)
	}()
}

// updateUIWithError updates the UI with an error message
func updateUIWithError(browser *ui.Browser, err error, url string) {
	log.Printf("Error loading page %s: %v", url, err)
	errorHTML := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Error</title></head>
		<body>
			<h1>Failed to load page</h1>
			<p>Could not load the page at %s.</p>
			<p>Error: %s</p>
		</body>
		</html>`, url, err.Error())
	_ = browser.RenderHTMLContent(errorHTML) // Ignore error for simplicity
	browser.HideLoading()
}

// updateUIWithContent updates the UI with HTML content
func updateUIWithContent(browser *ui.Browser, jsRuntime *js.Runtime, html string, url string) {
	log.Printf("Rendering page content")

	// Fyne widgets are thread-safe and can be updated from any goroutine
	// Render HTML using the canvas-based renderer
	err := browser.RenderHTMLContent(html)
	if err != nil {
		log.Printf("Error rendering HTML: %v", err)
		browser.SetContent("Error rendering HTML: " + err.Error())
		browser.HideLoading()
		return
	}

	log.Printf("Page loaded successfully")

	// Update tab title
	if title, ok := extractTitle(html); ok {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Goosie",
			Content: "Page loaded: " + title,
		})
		browser.UpdateActiveTabTitle(title)
	}

	// Hide loading indicator
	browser.HideLoading()

	// Set HTML content for JS runtime
	jsRuntime.SetHTMLContent(html)

	// Run any JavaScript on the page (optional)
	testScript := `console.log("Page loaded: " + document.title);`
	_, err = jsRuntime.RunScript(testScript)
	if err != nil {
		log.Printf("Error running JavaScript: %v", err)
	}
}

// loadPage fetches and displays a web page (deprecated - use loadPageAsync)
func loadPage(browser *ui.Browser, fetcher *net.Fetcher, parser *dom.Parser, jsRuntime *js.Runtime, url string) {
	loadPageAsync(browser, fetcher, parser, jsRuntime, url, context.Background())
}

// extractTitle parses the HTML and returns the content of the <title> tag.
func extractTitle(htmlContent string) (string, bool) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", false
	}

	var crawler func(*html.Node) (string, bool)
	crawler = func(node *html.Node) (string, bool) {
		if node.Type == html.ElementNode && node.Data == "title" {
			if node.FirstChild != nil {
				return node.FirstChild.Data, true
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if title, ok := crawler(c); ok {
				return title, ok
			}
		}
		return "", false
	}

	return crawler(doc)
}
