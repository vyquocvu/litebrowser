package renderer

import (
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
