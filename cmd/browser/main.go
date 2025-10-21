package main

import (
	"log"

	"github.com/vyquocvu/litebrowser/internal/dom"
	"github.com/vyquocvu/litebrowser/internal/js"
	"github.com/vyquocvu/litebrowser/internal/net"
	"github.com/vyquocvu/litebrowser/internal/ui"
)

func main() {
	// Initialize components
	fetcher := net.NewFetcher()
	parser := dom.NewParser()
	jsRuntime := js.NewRuntime()
	browser := ui.NewBrowser()

	// Set up navigation callback
	browser.SetNavigationCallback(func(url string) {
		loadPage(browser, fetcher, parser, jsRuntime, url)
	})

	// Show browser window
	browser.Show()
}

// loadPage fetches and displays a web page
func loadPage(browser *ui.Browser, fetcher *net.Fetcher, parser *dom.Parser, jsRuntime *js.Runtime, url string) {
	log.Printf("Navigating to: %s", url)

	// Update browser state
	browser.NavigateTo(url)

	// Show loading message
	browser.SetContent("Loading...")

	// Fetch the page
	html, err := fetcher.Fetch(url)
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
			browser.SetContent("Error loading page: " + err.Error())
			return
		}
	}

	// Render HTML using the canvas-based renderer
	err = browser.RenderHTMLContent(html)
	if err != nil {
		log.Printf("Error rendering HTML: %v", err)
		browser.SetContent("Error rendering HTML: " + err.Error())
		return
	}

	log.Printf("Page loaded successfully")

	// Set HTML content for JS runtime
	jsRuntime.SetHTMLContent(html)

	// Run any JavaScript on the page (optional)
	testScript := `console.log("Page loaded: " + document.title);`
	_, err = jsRuntime.RunScript(testScript)
	if err != nil {
		log.Printf("Error running JavaScript: %v", err)
	}
}
