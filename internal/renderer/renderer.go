package renderer

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	
	imageloader "github.com/vyquocvu/litebrowser/internal/image"
)

// NavigationCallback is called when a link is clicked
type NavigationCallback func(url string)

// Renderer is the main HTML renderer that coordinates parsing, layout, and rendering
type Renderer struct {
	layoutEngine   *LayoutEngine
	canvasRenderer *CanvasRenderer
	imageLoader    *imageloader.Loader

	// Cached trees for performance
	currentRenderTree *RenderNode
	currentLayoutTree *LayoutBox
	
	// Navigation callback for link clicks
	onNavigate NavigationCallback
	
	// Current page URL for resolving relative links
	currentURL string
}

// NewRenderer creates a new HTML renderer
func NewRenderer(width, height float32) *Renderer {
	imageLoader := imageloader.NewLoader(100) // Cache up to 100 images
	canvasRenderer := NewCanvasRenderer(width, height)
	canvasRenderer.imageLoader = imageLoader
	
	return &Renderer{
		layoutEngine:   NewLayoutEngine(width, height),
		canvasRenderer: canvasRenderer,
		imageLoader:    imageLoader,
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
	layoutTree := r.layoutEngine.ComputeLayout(renderTree)

	// Cache trees for viewport updates
	r.currentRenderTree = renderTree
	r.currentLayoutTree = layoutTree

	// Pass navigation callback to canvas renderer
	r.canvasRenderer.SetNavigationCallback(r.onNavigate, r.currentURL)

	// Render to canvas with viewport optimization
	canvasObject := r.canvasRenderer.RenderWithViewport(renderTree, layoutTree)

	return canvasObject, nil
}

// SetViewport updates the viewport for optimized rendering during scroll
func (r *Renderer) SetViewport(y, height float32) {
	r.canvasRenderer.SetViewport(y, height)
}

// UpdateViewport re-renders with the current viewport (for scroll updates)
func (r *Renderer) UpdateViewport() fyne.CanvasObject {
	if r.currentRenderTree == nil || r.currentLayoutTree == nil {
		return container.NewVBox()
	}
	return r.canvasRenderer.RenderWithViewport(r.currentRenderTree, r.currentLayoutTree)
}

// GetContentHeight returns the total height of the rendered content
func (r *Renderer) GetContentHeight() float32 {
	if r.currentLayoutTree == nil {
		return 0
	}
	return r.currentLayoutTree.Box.Height
}

// RenderHTMLBody renders just the body content of an HTML document
func (r *Renderer) RenderHTMLBody(htmlContent string) (fyne.CanvasObject, error) {
	// Use html.ParseFragment to handle content that is expected to be inside a <body> tag.
	// This avoids wrapping the content in an extra <html><body>...</body></html> structure.
	nodes, err := html.ParseFragment(strings.NewReader(htmlContent), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return nil, err
	}

	// Create a new root node to hold the parsed fragment.
	root := &html.Node{
		Type: html.ElementNode,
		Data: "body",
	}
	for _, node := range nodes {
		root.AppendChild(node)
	}

	// Build the render tree from the fragment.
	renderTree := BuildRenderTree(root)
	if renderTree == nil {
		return r.canvasRenderer.Render(nil), nil
	}

	// Perform layout.
	layoutTree := r.layoutEngine.ComputeLayout(renderTree)

	// Cache trees for viewport updates.
	r.currentRenderTree = renderTree
	r.currentLayoutTree = layoutTree

	// Pass navigation callback to canvas renderer.
	r.canvasRenderer.SetNavigationCallback(r.onNavigate, r.currentURL)

	// Render to canvas with viewport optimization.
	canvasObject := r.canvasRenderer.RenderWithViewport(renderTree, layoutTree)

	return canvasObject, nil
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

// SetNavigationCallback sets the callback for link clicks
func (r *Renderer) SetNavigationCallback(callback NavigationCallback) {
	r.onNavigate = callback
}

// SetCurrentURL sets the current page URL for resolving relative links
func (r *Renderer) SetCurrentURL(url string) {
	r.currentURL = url
}
