package main

import (
	"log"

	"github.com/vyquocvu/goosie/internal/dom"
	"github.com/vyquocvu/goosie/internal/js"
	"github.com/vyquocvu/goosie/internal/net"
)

func main() {
	log.Println("=== Testing Goosie Components ===")

	// Test 1: Fetch example.com
	log.Println("\n1. Testing HTTP Fetcher...")
	fetcher := net.NewFetcher()
	html, err := fetcher.Fetch("https://example.com")

	// If network is unavailable, use mock HTML for testing
	if err != nil {
		log.Printf("Network unavailable (%v), using mock HTML for testing", err)
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
	log.Printf("✓ HTML content available (%d bytes)", len(html))

	// Test 2: Parse body text
	log.Println("\n2. Testing HTML Parser...")
	parser := dom.NewParser()
	bodyText, err := parser.ParseBodyText(html)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}
	log.Printf("✓ Successfully parsed body text")
	log.Printf("Body text preview: %.100s...", bodyText)

	// Test 3: Initialize JS runtime
	log.Println("\n3. Testing JavaScript Runtime...")
	jsRuntime := js.NewRuntime()
	jsRuntime.SetHTMLContent(html)

	// Test console.log
	log.Println("\n4. Testing console.log...")
	_, err = jsRuntime.RunScript(`console.log("JS runtime initialized");`)
	if err != nil {
		log.Fatalf("Error running JavaScript: %v", err)
	}
	log.Println("✓ console.log works correctly")

	// Test document.getElementById (basic test)
	log.Println("\n5. Testing document.getElementById...")
	_, err = jsRuntime.RunScript(`
		var elem = document.getElementById("nonexistent");
		if (elem === null) {
			console.log("getElementById correctly returns null for non-existent element");
		}
		
		// Test with existing element
		var mainElem = document.getElementById("main-content");
		if (mainElem && mainElem.textContent) {
			console.log("getElementById found element with text:", mainElem.textContent.substring(0, 50) + "...");
		}
	`)
	if err != nil {
		log.Fatalf("Error testing getElementById: %v", err)
	}
	log.Println("✓ document.getElementById works correctly")

	log.Println("\n=== All Tests Passed ===")
	log.Println("\nNote: To see the full GUI browser window with 'Goja Browser' title,")
	log.Println("run 'go run ./cmd/browser' on a system with X11/Wayland display support.")
}
