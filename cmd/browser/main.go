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

	// Fetch example.com
	log.Println("Fetching https://example.com...")
	html, err := fetcher.Fetch("https://example.com")
	if err != nil {
		// Fallback to mock HTML if network is unavailable
		log.Printf("Network unavailable (%v), using mock HTML", err)
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
	}
	
	// Parse body text
	bodyText, err := parser.ParseBodyText(html)
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		browser.SetContent("Error parsing HTML: " + err.Error())
	} else {
		log.Printf("Parsed body text: %s", bodyText)
		browser.SetContent(bodyText)
	}

	// Set HTML content for JS runtime
	jsRuntime.SetHTMLContent(html)

	// Initialize JS runtime and run test script
	testScript := `console.log("JS runtime initialized");`
	log.Println("Running test JavaScript...")
	_, err = jsRuntime.RunScript(testScript)
	if err != nil {
		log.Printf("Error running JavaScript: %v", err)
	}

	// Show browser window
	browser.Show()
}
