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
	
	// nodeMap maps RenderNode IDs to their corresponding LayoutBoxes
	nodeMap map[int64]*LayoutBox
}

// NewLayoutEngine creates a new layout engine
func NewLayoutEngine(width, height float32) *LayoutEngine {
	return &LayoutEngine{
		canvasWidth:     width,
		canvasHeight:    height,
		defaultFontSize: 16.0,
		lineHeight:      1.5,
		nodeMap:         make(map[int64]*LayoutBox),
	}
}

// Layout performs layout calculations on the render tree and returns a layout tree
// This is the new API that produces a separate layout tree
func (le *LayoutEngine) ComputeLayout(root *RenderNode) *LayoutBox {
	if root == nil {
		return nil
	}
	
	// Clear previous mappings
	le.nodeMap = make(map[int64]*LayoutBox)
	
	// Build layout tree from render tree
	layoutRoot := le.buildLayoutBox(root, 0, 0, le.canvasWidth)
	
	return layoutRoot
}

// buildLayoutBox creates a LayoutBox for a RenderNode and computes its layout
func (le *LayoutEngine) buildLayoutBox(node *RenderNode, x, y, availableWidth float32) *LayoutBox {
	if node == nil {
		return nil
	}
	
	layoutBox := NewLayoutBox(node.ID)
	le.nodeMap[node.ID] = layoutBox
	
	// Determine display type
	if node.Type == NodeTypeElement {
		if node.IsBlock() {
			layoutBox.Display = DisplayBlock
		} else {
			layoutBox.Display = DisplayInline
		}
	} else {
		// Text nodes are inline
		layoutBox.Display = DisplayInline
	}
	
	// Compute layout
	currentY := le.computeLayoutBox(node, layoutBox, x, y, availableWidth)
	
	// Update height based on children
	layoutBox.Box.Height = currentY - y
	
	return layoutBox
}

// computeLayoutBox computes the layout for a single box
func (le *LayoutEngine) computeLayoutBox(node *RenderNode, layoutBox *LayoutBox, x, y, availableWidth float32) float32 {
	layoutBox.Box.X = x
	layoutBox.Box.Y = y
	layoutBox.Box.Width = availableWidth
	
	currentY := y
	
	if node.Type == NodeTypeText {
		// Layout text node
		currentY = le.computeTextLayout(node, layoutBox, x, y, availableWidth)
	} else if node.Type == NodeTypeElement {
		// Layout element node
		currentY = le.computeElementLayout(node, layoutBox, x, y, availableWidth)
	}
	
	return currentY
}

// computeTextLayout computes layout for text nodes
func (le *LayoutEngine) computeTextLayout(node *RenderNode, layoutBox *LayoutBox, x, y, availableWidth float32) float32 {
	fontSize := le.defaultFontSize
	
	// Calculate text dimensions (approximate)
	text := strings.TrimSpace(node.Text)
	if text == "" {
		layoutBox.Box.Height = 0
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
	
	layoutBox.Box.Height = textHeight
	
	return y + textHeight
}

// computeElementLayout computes layout for element nodes
func (le *LayoutEngine) computeElementLayout(node *RenderNode, layoutBox *LayoutBox, x, y, availableWidth float32) float32 {
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
			childLayoutBox := le.buildLayoutBox(child, x, childY, availableWidth)
			if childLayoutBox != nil {
				layoutBox.AddChild(childLayoutBox)
				childY = childLayoutBox.Box.Y + childLayoutBox.Box.Height
			}
		}
	} else {
		// Inline elements: layout children inline (simplified - just horizontal for now)
		childX := x
		for _, child := range node.Children {
			if child.Type == NodeTypeText {
				childLayoutBox := le.buildLayoutBox(child, childX, currentY, availableWidth-childX+x)
				if childLayoutBox != nil {
					layoutBox.AddChild(childLayoutBox)
					childY = childLayoutBox.Box.Y + childLayoutBox.Box.Height
				}
			} else {
				childLayoutBox := le.buildLayoutBox(child, childX, childY, availableWidth)
				if childLayoutBox != nil {
					layoutBox.AddChild(childLayoutBox)
					childY = childLayoutBox.Box.Y + childLayoutBox.Box.Height
				}
			}
		}
	}
	
	// Add bottom spacing for certain elements
	if verticalSpacing > 0 {
		childY += verticalSpacing
	}
	
	return childY
}

// GetLayoutBox returns the LayoutBox for a given RenderNode ID
func (le *LayoutEngine) GetLayoutBox(nodeID int64) *LayoutBox {
	return le.nodeMap[nodeID]
}

// HitTest performs hit testing on the layout tree
// Returns the node ID of the deepest layout box containing the point (x, y)
// Returns 0 if no box contains the point
func (le *LayoutEngine) HitTest(layoutRoot *LayoutBox, x, y float32) int64 {
	if layoutRoot == nil {
		return 0
	}
	
	return le.hitTestRecursive(layoutRoot, x, y)
}

// hitTestRecursive recursively searches for the deepest box containing (x, y)
func (le *LayoutEngine) hitTestRecursive(box *LayoutBox, x, y float32) int64 {
	if !box.Contains(x, y) {
		return 0
	}
	
	// Check children first (depth-first search for deepest match)
	for _, child := range box.Children {
		if hitID := le.hitTestRecursive(child, x, y); hitID != 0 {
			return hitID
		}
	}
	
	// If no child contains the point, return this box's node ID
	return box.NodeID
}

// Layout performs layout calculations on the render tree (deprecated - use ComputeLayout)
// Kept for backward compatibility
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
