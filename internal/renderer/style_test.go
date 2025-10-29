package renderer

import (
	"image/color"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestStyleApplication(t *testing.T) {
	htmlContent := `
		<html>
			<head>
				<style>
					h1 {
						display: block;
						font-size: 32px;
					}
				</style>
			</head>
			<body>
				<h1>Hello, world!</h1>
			</body>
		</html>
	`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}

	stylesheet := extractAndParseCSS(doc)
	if len(stylesheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(stylesheet.Rules))
	}

	renderTree := BuildRenderTree(findBodyNode(doc))
	if renderTree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}

	styleManager := NewStyleManager(stylesheet)
	styleManager.ApplyStyles(renderTree)

	h1Node := findNodeByTag(renderTree, "h1")
	if h1Node == nil {
		t.Fatal("h1 node not found in render tree")
	}

	if h1Node.ComputedStyle.Display != "block" {
		t.Errorf("expected display 'block', got '%s'", h1Node.ComputedStyle.Display)
	}
	if h1Node.ComputedStyle.FontSize != 32.0 {
		t.Errorf("expected font-size 32.0, got %f", h1Node.ComputedStyle.FontSize)
	}
}

func TestAdvancedStyleApplication(t *testing.T) {
	htmlContent := `
		<html>
			<head>
				<style>
					body {
						font-size: 16px;
						background-color: #eee;
						width: 60vw;
						margin: 15vh auto;
						font-family: system-ui, sans-serif;
					}
					h1 { font-size: 1.5em; }
					div { opacity: 0.8; }
					a:link { color: #348; }
				</style>
			</head>
			<body>
				<h1>Title</h1>
				<div>A div</div>
				<a href="#">Link</a>
			</body>
		</html>
	`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}

	stylesheet := extractAndParseCSS(doc)
	renderTree := BuildRenderTree(findBodyNode(doc))
	if renderTree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}

	styleManager := NewStyleManager(stylesheet)
	styleManager.ApplyStyles(renderTree)

	bodyNode := findNodeByTag(renderTree, "body")
	if bodyNode == nil {
		t.Fatal("body node not found in render tree")
	}
	expectedBgColor := color.RGBA{R: 0xee, G: 0xee, B: 0xee, A: 0xff}
	if bodyNode.ComputedStyle.BackgroundColor != expectedBgColor {
		t.Errorf("expected background color %v, got %v", expectedBgColor, bodyNode.ComputedStyle.BackgroundColor)
	}
	if bodyNode.ComputedStyle.Width != "60vw" {
		t.Errorf("expected width '60vw', got '%s'", bodyNode.ComputedStyle.Width)
	}
	if bodyNode.ComputedStyle.Margin != "15vh auto" {
		t.Errorf("expected margin '15vh auto', got '%s'", bodyNode.ComputedStyle.Margin)
	}
	if bodyNode.ComputedStyle.FontFamily != "system-ui, sans-serif" {
		t.Errorf("expected font-family 'system-ui, sans-serif', got '%s'", bodyNode.ComputedStyle.FontFamily)
	}

	h1Node := findNodeByTag(renderTree, "h1")
	if h1Node == nil {
		t.Fatal("h1 node not found in render tree")
	}
	expectedFontSize := float32(24.0)
	if h1Node.ComputedStyle.FontSize != expectedFontSize {
		t.Errorf("expected font-size %f, got %f", expectedFontSize, h1Node.ComputedStyle.FontSize)
	}

	divNode := findNodeByTag(renderTree, "div")
	if divNode == nil {
		t.Fatal("div node not found in render tree")
	}
	if divNode.ComputedStyle.Opacity != 0.8 {
		t.Errorf("expected opacity 0.8, got %f", divNode.ComputedStyle.Opacity)
	}

	aNode := findNodeByTag(renderTree, "a")
	if aNode == nil {
		t.Fatal("a node not found in render tree")
	}
	expectedLinkColor := color.RGBA{R: 0x33, G: 0x44, B: 0x88, A: 0xff}
	if aNode.ComputedStyle.Color != expectedLinkColor {
		t.Errorf("expected color %v, got %v", expectedLinkColor, aNode.ComputedStyle.Color)
	}
}

func findNodeByTag(node *RenderNode, tagName string) *RenderNode {
	if node.TagName == tagName {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeByTag(child, tagName); found != nil {
			return found
		}
	}
	return nil
}

func TestNamedColorApplication(t *testing.T) {
	htmlContent := `
		<html>
			<head>
				<style>
					div {
						color: red;
						background-color: blue;
					}
				</style>
			</head>
			<body>
				<div>Red text, blue background</div>
			</body>
		</html>
	`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}

	stylesheet := extractAndParseCSS(doc)
	renderTree := BuildRenderTree(findBodyNode(doc))
	if renderTree == nil {
		t.Fatal("BuildRenderTree returned nil")
	}

	styleManager := NewStyleManager(stylesheet)
	styleManager.ApplyStyles(renderTree)

	divNode := findNodeByTag(renderTree, "div")
	if divNode == nil {
		t.Fatal("div node not found in render tree")
	}

	expectedColor := color.RGBA{R: 0xff, A: 0xff}
	if divNode.ComputedStyle.Color != expectedColor {
		t.Errorf("expected color %v, got %v", expectedColor, divNode.ComputedStyle.Color)
	}

	expectedBgColor := color.RGBA{B: 0xff, A: 0xff}
	if divNode.ComputedStyle.BackgroundColor != expectedBgColor {
		t.Errorf("expected background color %v, got %v", expectedBgColor, divNode.ComputedStyle.BackgroundColor)
	}
}
