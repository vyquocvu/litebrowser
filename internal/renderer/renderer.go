package renderer

import (
	"strings"
	
	"fyne.io/fyne/v2"
	"golang.org/x/net/html"
)

// Renderer is the main HTML renderer that coordinates parsing, layout, and rendering
type Renderer struct {
	layoutEngine   *LayoutEngine
	canvasRenderer *CanvasRenderer
}

// NewRenderer creates a new HTML renderer
func NewRenderer(width, height float32) *Renderer {
	return &Renderer{
		layoutEngine:   NewLayoutEngine(width, height),
		canvasRenderer: NewCanvasRenderer(width, height),
	}
}

// RenderHTML renders HTML content and returns a Fyne canvas object
func (r *Renderer) RenderHTML(htmlContent string) (fyne.CanvasObject, error) {
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}
	
	// Find body element
	bodyNode := findBodyNode(doc)
	if bodyNode == nil {
		// No body found, use the entire document
		bodyNode = doc
	}
	
	// Build render tree
	renderTree := BuildRenderTree(bodyNode)
	if renderTree == nil {
		// Return empty container if no content
		return r.canvasRenderer.Render(nil), nil
	}
	
	// Perform layout
	r.layoutEngine.Layout(renderTree)
	
	// Render to canvas
	canvasObject := r.canvasRenderer.Render(renderTree)
	
	return canvasObject, nil
}

// RenderHTMLBody renders just the body content of an HTML document
func (r *Renderer) RenderHTMLBody(htmlContent string) (fyne.CanvasObject, error) {
	return r.RenderHTML(htmlContent)
}

// findBodyNode finds the body element in an HTML document
func findBodyNode(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	
	if node.Type == html.ElementNode && node.Data == "body" {
		return node
	}
	
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findBodyNode(child); found != nil {
			return found
		}
	}
	
	return nil
}

// SetSize updates the renderer dimensions
func (r *Renderer) SetSize(width, height float32) {
	r.layoutEngine.canvasWidth = width
	r.layoutEngine.canvasHeight = height
	r.canvasRenderer.canvasWidth = width
	r.canvasRenderer.canvasHeight = height
}
