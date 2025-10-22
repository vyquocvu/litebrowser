package renderer

import (
	"image/color"
	"strings"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CanvasRenderer renders a render tree onto a Fyne canvas
type CanvasRenderer struct {
	canvasWidth  float32
	canvasHeight float32
	defaultSize  float32
	
	// Viewport for optimized rendering
	viewportY      float32
	viewportHeight float32
	
	// Cached display list for performance
	cachedDisplayList *DisplayList
	cachedLayoutRoot  *LayoutBox
	cachedRenderRoot  *RenderNode
}

// NewCanvasRenderer creates a new canvas renderer
func NewCanvasRenderer(width, height float32) *CanvasRenderer {
	return &CanvasRenderer{
		canvasWidth:    width,
		canvasHeight:   height,
		defaultSize:    16.0,
		viewportY:      0,
		viewportHeight: height,
	}
}

// SetViewport sets the current viewport for optimized rendering
func (cr *CanvasRenderer) SetViewport(y, height float32) {
	cr.viewportY = y
	cr.viewportHeight = height
}

// isInViewport checks if a box intersects with the current viewport
func (cr *CanvasRenderer) isInViewport(box Rect) bool {
	// Add buffer zone above and below viewport for smoother scrolling
	bufferZone := cr.viewportHeight * 0.5
	viewportTop := cr.viewportY - bufferZone
	viewportBottom := cr.viewportY + cr.viewportHeight + bufferZone
	
	boxBottom := box.Y + box.Height
	
	// Check if box intersects with viewport
	return boxBottom >= viewportTop && box.Y <= viewportBottom
}

// Render renders the render tree and returns a Fyne container
func (cr *CanvasRenderer) Render(root *RenderNode) fyne.CanvasObject {
	if root == nil {
		return container.NewVBox()
	}
	
	objects := make([]fyne.CanvasObject, 0)
	cr.renderNode(root, &objects)
	
	return container.NewVBox(objects...)
}

// renderNode renders a single node and its children
func (cr *CanvasRenderer) renderNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	if node == nil {
		return
	}
	
	if node.Type == NodeTypeText {
		cr.renderTextNode(node, objects)
	} else if node.Type == NodeTypeElement {
		cr.renderElementNode(node, objects)
	}
}

// renderTextNode renders a text node
func (cr *CanvasRenderer) renderTextNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := strings.TrimSpace(node.Text)
	if text == "" {
		return
	}
	
	// Create text widget
	textWidget := widget.NewLabel(text)
	textWidget.Wrapping = fyne.TextWrapWord
	
	// Get text style from parent if available
	if node.Parent != nil {
		textWidget.TextStyle = cr.getTextStyle(node.Parent.TagName)
	}
	
	*objects = append(*objects, textWidget)
}

// renderElementNode renders an element node
func (cr *CanvasRenderer) renderElementNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	switch node.TagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		cr.renderHeading(node, objects)
	case "p":
		cr.renderParagraph(node, objects)
	case "div":
		cr.renderDiv(node, objects)
	case "a":
		cr.renderLink(node, objects)
	case "ul", "ol":
		cr.renderList(node, objects)
	case "li":
		cr.renderListItem(node, objects)
	case "img":
		cr.renderImage(node, objects)
	case "br":
		// Add a spacer for line break
		*objects = append(*objects, widget.NewLabel(""))
	case "span", "strong", "em", "b", "i":
		// Inline elements - render children
		for _, child := range node.Children {
			cr.renderNode(child, objects)
		}
	default:
		// Generic element - just render children
		for _, child := range node.Children {
			cr.renderNode(child, objects)
		}
	}
}

// renderHeading renders heading elements (h1-h6)
func (cr *CanvasRenderer) renderHeading(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}
	
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle = fyne.TextStyle{Bold: true}
	
	// Different sizes for different heading levels
	// Note: Fyne doesn't support arbitrary font sizes directly,
	// so we use TextStyle to make headings bold
	
	*objects = append(*objects, label)
}

// renderParagraph renders paragraph elements
func (cr *CanvasRenderer) renderParagraph(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}
	
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	
	*objects = append(*objects, label)
}

// renderDiv renders div elements
func (cr *CanvasRenderer) renderDiv(node *RenderNode, objects *[]fyne.CanvasObject) {
	// Render children
	for _, child := range node.Children {
		cr.renderNode(child, objects)
	}
}

// renderLink renders anchor (link) elements
func (cr *CanvasRenderer) renderLink(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	href, hasHref := node.GetAttribute("href")
	
	if text == "" {
		return
	}
	
	if hasHref && href != "" {
		// Create a hyperlink widget
		link := widget.NewHyperlink(text, nil)
		// Note: Fyne's hyperlink requires a proper URL parse, 
		// but for now we'll just display as styled text
		*objects = append(*objects, link)
	} else {
		// No href, just display as text
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		*objects = append(*objects, label)
	}
}

// renderList renders ul/ol elements
func (cr *CanvasRenderer) renderList(node *RenderNode, objects *[]fyne.CanvasObject) {
	// Render list items
	for _, child := range node.Children {
		if child.TagName == "li" {
			cr.renderListItem(child, objects)
		}
	}
}

// renderListItem renders li elements
func (cr *CanvasRenderer) renderListItem(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}
	
	// Add bullet point
	label := widget.NewLabel("• " + text)
	label.Wrapping = fyne.TextWrapWord
	
	*objects = append(*objects, label)
}

// renderImage renders img elements
func (cr *CanvasRenderer) renderImage(node *RenderNode, objects *[]fyne.CanvasObject) {
	alt, hasAlt := node.GetAttribute("alt")
	src, hasSrc := node.GetAttribute("src")
	
	// For now, just display alt text or placeholder
	// Full image loading would require fetching the image data
	displayText := "[Image"
	if hasSrc {
		displayText += ": " + src
	}
	if hasAlt {
		displayText += " - " + alt
	}
	displayText += "]"
	
	label := widget.NewLabel(displayText)
	label.Wrapping = fyne.TextWrapWord
	
	// Create a colored rectangle to represent the image placeholder
	rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	rect.SetMinSize(fyne.NewSize(100, 100))
	
	*objects = append(*objects, container.NewVBox(rect, label))
}

// extractText extracts all text content from a node and its children
func (cr *CanvasRenderer) extractText(node *RenderNode) string {
	var text strings.Builder
	cr.extractTextRecursive(node, &text)
	return strings.TrimSpace(text.String())
}

// extractTextRecursive recursively extracts text from a node tree
func (cr *CanvasRenderer) extractTextRecursive(node *RenderNode, builder *strings.Builder) {
	if node == nil {
		return
	}
	
	if node.Type == NodeTypeText {
		text := strings.TrimSpace(node.Text)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(text)
		}
	}
	
	for _, child := range node.Children {
		cr.extractTextRecursive(child, builder)
	}
}

// getFontSize returns font size for an element type
func (cr *CanvasRenderer) getFontSize(tagName string) float32 {
	fontSizes := map[string]float32{
		"h1": cr.defaultSize * 2.0,
		"h2": cr.defaultSize * 1.5,
		"h3": cr.defaultSize * 1.17,
		"h4": cr.defaultSize * 1.0,
		"h5": cr.defaultSize * 0.83,
		"h6": cr.defaultSize * 0.67,
		"p":  cr.defaultSize,
	}
	
	if size, ok := fontSizes[tagName]; ok {
		return size
	}
	return cr.defaultSize
}

// getTextStyle returns text style for an element type
func (cr *CanvasRenderer) getTextStyle(tagName string) fyne.TextStyle {
	switch tagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return fyne.TextStyle{Bold: true}
	case "strong", "b":
		return fyne.TextStyle{Bold: true}
	case "em", "i":
		return fyne.TextStyle{Italic: true}
	default:
		return fyne.TextStyle{}
	}
}

// RenderWithViewport renders the render tree with viewport culling for better performance
func (cr *CanvasRenderer) RenderWithViewport(root *RenderNode, layoutRoot *LayoutBox) fyne.CanvasObject {
	if root == nil || layoutRoot == nil {
		return container.NewVBox()
	}
	
	// Build or reuse display list
	var displayList *DisplayList
	if cr.cachedDisplayList != nil && cr.cachedRenderRoot == root && cr.cachedLayoutRoot == layoutRoot {
		// Reuse cached display list
		displayList = cr.cachedDisplayList
	} else {
		// Build new display list
		dlb := NewDisplayListBuilder()
		displayList = dlb.Build(layoutRoot, root)
		
		// Cache for next time
		cr.cachedDisplayList = displayList
		cr.cachedRenderRoot = root
		cr.cachedLayoutRoot = layoutRoot
	}
	
	// Filter commands based on viewport
	objects := make([]fyne.CanvasObject, 0)
	for _, cmd := range displayList.Commands {
		if cr.isInViewport(cmd.Box) {
			cr.renderCommand(cmd, &objects)
		}
	}
	
	if len(objects) == 0 {
		return container.NewVBox()
	}
	
	return container.NewVBox(objects...)
}

// renderCommand renders a single paint command to canvas objects
func (cr *CanvasRenderer) renderCommand(cmd *PaintCommand, objects *[]fyne.CanvasObject) {
	switch cmd.Type {
	case PaintText:
		if strings.TrimSpace(cmd.Text) == "" {
			return
		}
		
		label := widget.NewLabel(cmd.Text)
		label.Wrapping = fyne.TextWrapWord
		
		if cmd.Bold && cmd.Italic {
			label.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
		} else if cmd.Bold {
			label.TextStyle = fyne.TextStyle{Bold: true}
		} else if cmd.Italic {
			label.TextStyle = fyne.TextStyle{Italic: true}
		}
		
		*objects = append(*objects, label)
		
	case PaintRect:
		rect := canvas.NewRectangle(cmd.FillColor)
		rect.SetMinSize(fyne.NewSize(cmd.Box.Width, cmd.Box.Height))
		*objects = append(*objects, rect)
		
	case PaintImage:
		// Render image placeholder
		displayText := "[Image"
		if cmd.ImageSrc != "" {
			displayText += ": " + cmd.ImageSrc
		}
		if cmd.ImageAlt != "" {
			displayText += " - " + cmd.ImageAlt
		}
		displayText += "]"
		
		label := widget.NewLabel(displayText)
		label.Wrapping = fyne.TextWrapWord
		
		rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		rect.SetMinSize(fyne.NewSize(100, 100))
		
		*objects = append(*objects, container.NewVBox(rect, label))
	}
}

// ClearCache clears the cached display list to force re-rendering
func (cr *CanvasRenderer) ClearCache() {
	cr.cachedDisplayList = nil
	cr.cachedLayoutRoot = nil
	cr.cachedRenderRoot = nil
}
