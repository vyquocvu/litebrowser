package renderer

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

func TestDebugRenderTree(t *testing.T) {
	// This is the EXACT HTML from the bug report, including the malformed parts
	htmlContent := `<!doctype html>
<html lang="en">
<head>
  <title>Example Domain</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body{background:#eee;width:60vw;margin:15vh auto;font-family:system-ui,sans-serif}
    h1{font-size:1.5em}
    div{opacity:0.8}
    a:link,a:visited{color:#348}
  </style>
<body>
  <div>
    <h1>Example Domain</h1>
    <p>This domain is for use in documentation examples without needing permission. Avoid use in operations.
    <p><a href="https://iana.org/domains/example">Learn more</a>
  </div>
</body>
</html>`

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Error parsing HTML: %v", err)
	}

	// Print the entire document structure to understand the parsing
	fmt.Println("=== Full Document Structure ===")
	printHTMLNode(doc, 0)

	// Find body element
	bodyNode := findBodyNode(doc)
	if bodyNode == nil {
		t.Fatal("No body found")
	}

	// Count how many body nodes exist
	bodyCount := 0
	var countBodies func(*html.Node)
	countBodies = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyCount++
			fmt.Printf("Found body node %d at depth\n", bodyCount)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			countBodies(child)
		}
	}
	countBodies(doc)
	fmt.Printf("\nTotal body nodes found: %d\n\n", bodyCount)

	// Print the HTML structure
	fmt.Println("=== Body HTML Structure ===")
	printHTMLNode(bodyNode, 0)

	// Build render tree
	renderTree := BuildRenderTree(bodyNode)
	if renderTree == nil {
		t.Fatal("No render tree created")
	}

	// Print the render tree structure
	fmt.Println("\n=== Render Tree ===")
	printRenderTree(renderTree, 0)
}

func printHTMLNode(node *html.Node, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	
	switch node.Type {
	case html.ElementNode:
		fmt.Printf("%s<%s> (type=%v)\n", indent, node.Data, node.Type)
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			fmt.Printf("%sText: %q\n", indent, text)
		}
	case html.DocumentNode:
		fmt.Printf("%sDocument\n", indent)
	case html.DoctypeNode:
		fmt.Printf("%sDoctype: %s\n", indent, node.Data)
	case html.CommentNode:
		fmt.Printf("%sComment: %q\n", indent, node.Data)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		printHTMLNode(child, depth+1)
	}
}

func printRenderTree(node *RenderNode, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	
	if node.Type == NodeTypeText {
		fmt.Printf("%sText: %q\n", indent, node.Text)
	} else {
		fmt.Printf("%s<%s>\n", indent, node.TagName)
	}

	for _, child := range node.Children {
		printRenderTree(child, depth+1)
	}
}

func printLayoutTree(box *LayoutBox, depth int) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sLayoutBox (NodeID=%d, Display=%v)\n", indent, box.NodeID, box.Display)
	
	for _, child := range box.Children {
		printLayoutTree(child, depth+1)
	}
}

func TestCountRenderedObjects(t *testing.T) {
	htmlContent := `<!doctype html>
<html lang="en">
<head>
  <title>Example Domain</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body{background:#eee;width:60vw;margin:15vh auto;font-family:system-ui,sans-serif}
    h1{font-size:1.5em}
    div{opacity:0.8}
    a:link,a:visited{color:#348}
  </style>
<body>
  <div>
    <h1>Example Domain</h1>
    <p>This domain is for use in documentation examples without needing permission. Avoid use in operations.
    <p><a href="https://iana.org/domains/example">Learn more</a>
  </div>
</body>
</html>`

	// Create renderer and render the HTML
	htmlRenderer := NewRenderer(800, 600)
	
	// Let's also check the display list
	doc, _ := html.Parse(strings.NewReader(htmlContent))
	bodyNode := findBodyNode(doc)
	renderTree := BuildRenderTree(bodyNode)
	
	// Perform layout
	layoutTree := htmlRenderer.layoutEngine.ComputeLayout(renderTree)
	
	// Print layout tree
	fmt.Println("=== Layout Tree ===")
	printLayoutTree(layoutTree, 0)
	
	// Build display list manually to inspect it
	dlb := NewDisplayListBuilder()
	displayList := dlb.Build(layoutTree, renderTree)
	
	fmt.Printf("\nDisplay list has %d commands:\n", len(displayList.Commands))
	textCommands := 0
	for i, cmd := range displayList.Commands {
		if cmd.Type == PaintText {
			textCommands++
			fmt.Printf("  Command %d: Text=%q\n", i, cmd.Text)
		}
	}
	fmt.Printf("Total text commands: %d\n\n", textCommands)
	
	canvasObject, err := htmlRenderer.RenderHTML(htmlContent)
	if err != nil {
		t.Fatalf("Error rendering HTML: %v", err)
	}

	// Count the number of objects
	vbox := canvasObject.(*fyne.Container)
	fmt.Printf("Number of rendered objects: %d\n", len(vbox.Objects))
	for i, obj := range vbox.Objects {
		fmt.Printf("  Object %d: %T\n", i, obj)
		if label, isLabel := obj.(*widget.Label); isLabel {
			fmt.Printf("    Text: %q\n", label.Text)
		}
	}
}
