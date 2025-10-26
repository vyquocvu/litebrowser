package renderer

import (
	"strings"
	"testing"
	
	"golang.org/x/net/html"
)

func TestRenderCodeElement(t *testing.T) {
	htmlContent := `<code>const x = 42;</code>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the code element
	var codeHTMLNode *html.Node
	var findCode func(*html.Node)
	findCode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "code" {
			codeHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCode(c)
		}
	}
	findCode(doc)
	
	if codeHTMLNode == nil {
		t.Fatal("Code HTML node not found")
	}
	
	tree := BuildRenderTree(codeHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	if tree.TagName != "code" {
		t.Errorf("Expected tag name 'code', got '%s'", tree.TagName)
	}
}

func TestRenderPreElement(t *testing.T) {
	htmlContent := `<pre>  Line 1
  Line 2
    Indented</pre>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the pre element
	var preHTMLNode *html.Node
	var findPre func(*html.Node)
	findPre = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			preHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findPre(c)
		}
	}
	findPre(doc)
	
	if preHTMLNode == nil {
		t.Fatal("Pre HTML node not found")
	}
	
	tree := BuildRenderTree(preHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	if tree.TagName != "pre" {
		t.Errorf("Expected tag name 'pre', got '%s'", tree.TagName)
	}
	
	// Pre should be a block element
	if !tree.IsBlock() {
		t.Error("Pre element should be a block element")
	}
}

func TestRenderBlockquoteElement(t *testing.T) {
	htmlContent := `<blockquote>This is a quote</blockquote>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the blockquote element
	var blockquoteHTMLNode *html.Node
	var findBlockquote func(*html.Node)
	findBlockquote = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "blockquote" {
			blockquoteHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBlockquote(c)
		}
	}
	findBlockquote(doc)
	
	if blockquoteHTMLNode == nil {
		t.Fatal("Blockquote HTML node not found")
	}
	
	tree := BuildRenderTree(blockquoteHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	if tree.TagName != "blockquote" {
		t.Errorf("Expected tag name 'blockquote', got '%s'", tree.TagName)
	}
	
	// Blockquote should be a block element
	if !tree.IsBlock() {
		t.Error("Blockquote element should be a block element")
	}
}

func TestFontMetricsCodeStyle(t *testing.T) {
	fm := NewFontMetrics(16.0)
	
	// Test code element style
	codeStyle := fm.GetTextStyle("code")
	if !codeStyle.Monospace {
		t.Error("Code element should have monospace style")
	}
	
	// Test pre element style
	preStyle := fm.GetTextStyle("pre")
	if !preStyle.Monospace {
		t.Error("Pre element should have monospace style")
	}
}

func TestCanvasRendererCodeElements(t *testing.T) {
	cr := NewCanvasRenderer(800, 600)
	
	// Test rendering code element
	htmlContent := `<code>const x = 42;</code>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the code element
	var codeHTMLNode *html.Node
	var findCode func(*html.Node)
	findCode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "code" {
			codeHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCode(c)
		}
	}
	findCode(doc)
	
	if codeHTMLNode == nil {
		t.Fatal("Code HTML node not found")
	}
	
	tree := BuildRenderTree(codeHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	// Render should not panic
	obj := cr.Render(tree)
	if obj == nil {
		t.Error("Render returned nil for code element")
	}
}

func TestCanvasRendererPreElements(t *testing.T) {
	cr := NewCanvasRenderer(800, 600)
	
	// Test rendering pre element
	htmlContent := `<pre>Line 1
Line 2</pre>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the pre element
	var preHTMLNode *html.Node
	var findPre func(*html.Node)
	findPre = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			preHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findPre(c)
		}
	}
	findPre(doc)
	
	if preHTMLNode == nil {
		t.Fatal("Pre HTML node not found")
	}
	
	tree := BuildRenderTree(preHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	// Render should not panic
	obj := cr.Render(tree)
	if obj == nil {
		t.Error("Render returned nil for pre element")
	}
}

func TestCanvasRendererBlockquoteElements(t *testing.T) {
	cr := NewCanvasRenderer(800, 600)
	
	// Test rendering blockquote element
	htmlContent := `<blockquote>This is a quote</blockquote>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	// Find the blockquote element
	var blockquoteHTMLNode *html.Node
	var findBlockquote func(*html.Node)
	findBlockquote = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "blockquote" {
			blockquoteHTMLNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBlockquote(c)
		}
	}
	findBlockquote(doc)
	
	if blockquoteHTMLNode == nil {
		t.Fatal("Blockquote HTML node not found")
	}
	
	tree := BuildRenderTree(blockquoteHTMLNode)
	if tree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}
	
	// Render should not panic
	obj := cr.Render(tree)
	if obj == nil {
		t.Error("Render returned nil for blockquote element")
	}
}
