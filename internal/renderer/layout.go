package renderer

import (
	"strings"
)

// LayoutEngine handles layout calculations for render nodes
type LayoutEngine struct {
	canvasWidth  float32
	canvasHeight float32
	
	// Default font sizes for headings and text
	defaultFontSize float32
	lineHeight      float32
}

// NewLayoutEngine creates a new layout engine
func NewLayoutEngine(width, height float32) *LayoutEngine {
	return &LayoutEngine{
		canvasWidth:     width,
		canvasHeight:    height,
		defaultFontSize: 16.0,
		lineHeight:      1.5,
	}
}

// Layout performs layout calculations on the render tree
func (le *LayoutEngine) Layout(root *RenderNode) {
	if root == nil {
		return
	}
	
	// Start layout from top-left with full canvas width
	le.layoutNode(root, 0, 0, le.canvasWidth)
}

// layoutNode performs layout calculation for a single node and its children
func (le *LayoutEngine) layoutNode(node *RenderNode, x, y, availableWidth float32) float32 {
	if node == nil {
		return y
	}
	
	currentY := y
	
	if node.Type == NodeTypeText {
		// Layout text node
		currentY = le.layoutTextNode(node, x, y, availableWidth)
	} else if node.Type == NodeTypeElement {
		// Layout element node
		currentY = le.layoutElementNode(node, x, y, availableWidth)
	}
	
	return currentY
}

// layoutTextNode handles layout for text nodes
func (le *LayoutEngine) layoutTextNode(node *RenderNode, x, y, availableWidth float32) float32 {
	fontSize := le.defaultFontSize
	
	// Calculate text dimensions (approximate)
	text := strings.TrimSpace(node.Text)
	if text == "" {
		return y
	}
	
	// Approximate character width (varies by font)
	charWidth := fontSize * 0.6
	charsPerLine := int(availableWidth / charWidth)
	
	if charsPerLine < 1 {
		charsPerLine = 1
	}
	
	// Calculate number of lines needed
	lines := (len(text) + charsPerLine - 1) / charsPerLine
	textHeight := float32(lines) * fontSize * le.lineHeight
	
	node.Box.X = x
	node.Box.Y = y
	node.Box.Width = availableWidth
	node.Box.Height = textHeight
	
	return y + textHeight
}

// layoutElementNode handles layout for element nodes
func (le *LayoutEngine) layoutElementNode(node *RenderNode, x, y, availableWidth float32) float32 {
	// Set initial position
	node.Box.X = x
	node.Box.Y = y
	node.Box.Width = availableWidth
	
	// Calculate spacing based on element type
	verticalSpacing := le.getVerticalSpacing(node.TagName)
	
	currentY := y
	
	// Add top spacing for certain elements
	if verticalSpacing > 0 {
		currentY += verticalSpacing
	}
	
	// Layout children
	childY := currentY
	
	if node.IsBlock() {
		// Block elements: stack children vertically
		for _, child := range node.Children {
			childY = le.layoutNode(child, x, childY, availableWidth)
		}
	} else {
		// Inline elements: layout children inline (simplified - just horizontal for now)
		childX := x
		for _, child := range node.Children {
			if child.Type == NodeTypeText {
				childY = le.layoutTextNode(child, childX, currentY, availableWidth-childX+x)
				// For inline layout, we'd advance childX here
				// For simplicity, we're just doing basic vertical stacking
			} else {
				childY = le.layoutNode(child, childX, childY, availableWidth)
			}
		}
	}
	
	// Calculate total height
	node.Box.Height = childY - currentY
	
	// Add bottom spacing for certain elements
	if verticalSpacing > 0 {
		childY += verticalSpacing
	}
	
	return childY
}

// getFontSize returns the font size for an element
func (le *LayoutEngine) getFontSize(tagName string) float32 {
	fontSizes := map[string]float32{
		"h1": le.defaultFontSize * 2.0,
		"h2": le.defaultFontSize * 1.5,
		"h3": le.defaultFontSize * 1.17,
		"h4": le.defaultFontSize * 1.0,
		"h5": le.defaultFontSize * 0.83,
		"h6": le.defaultFontSize * 0.67,
		"p":  le.defaultFontSize,
	}
	
	if size, ok := fontSizes[tagName]; ok {
		return size
	}
	return le.defaultFontSize
}

// getVerticalSpacing returns the vertical spacing (margin) for an element
func (le *LayoutEngine) getVerticalSpacing(tagName string) float32 {
	spacing := map[string]float32{
		"h1": le.defaultFontSize * 0.67,
		"h2": le.defaultFontSize * 0.67,
		"h3": le.defaultFontSize * 0.67,
		"h4": le.defaultFontSize * 0.67,
		"h5": le.defaultFontSize * 0.67,
		"h6": le.defaultFontSize * 0.67,
		"p":  le.defaultFontSize * 0.5,
		"ul": le.defaultFontSize * 0.5,
		"ol": le.defaultFontSize * 0.5,
		"li": le.defaultFontSize * 0.25,
	}
	
	if s, ok := spacing[tagName]; ok {
		return s
	}
	return 0
}
