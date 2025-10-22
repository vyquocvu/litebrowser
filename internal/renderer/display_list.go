package renderer

import (
	"image/color"
)

// PaintCommandType represents the type of paint command
type PaintCommandType int

const (
	// PaintText represents a text paint command
	PaintText PaintCommandType = iota
	// PaintRect represents a rectangle paint command
	PaintRect
	// PaintImage represents an image paint command
	PaintImage
)

// PaintCommand represents a single paint operation
type PaintCommand struct {
	Type   PaintCommandType
	NodeID int64   // ID of the node this command is for
	Box    Rect    // Position and size for the command
	
	// Text-specific fields
	Text      string
	FontSize  float32
	Bold      bool
	Italic    bool
	
	// Rectangle-specific fields
	FillColor   color.Color
	StrokeColor color.Color
	StrokeWidth float32
	
	// Image-specific fields
	ImageSrc string
	ImageAlt string
}

// DisplayList represents a list of paint commands
type DisplayList struct {
	Commands []*PaintCommand
}

// NewDisplayList creates a new display list
func NewDisplayList() *DisplayList {
	return &DisplayList{
		Commands: make([]*PaintCommand, 0),
	}
}

// AddCommand adds a paint command to the display list
func (dl *DisplayList) AddCommand(cmd *PaintCommand) {
	dl.Commands = append(dl.Commands, cmd)
}

// Clear removes all commands from the display list
func (dl *DisplayList) Clear() {
	dl.Commands = make([]*PaintCommand, 0)
}

// DisplayListBuilder builds a display list from a layout tree and render tree
type DisplayListBuilder struct {
	defaultFontSize float32
	fontMetrics     *FontMetrics
}

// NewDisplayListBuilder creates a new display list builder
func NewDisplayListBuilder() *DisplayListBuilder {
	defaultSize := float32(16.0)
	return &DisplayListBuilder{
		defaultFontSize: defaultSize,
		fontMetrics:     NewFontMetrics(defaultSize),
	}
}

// Build builds a display list from a layout tree and render tree
func (dlb *DisplayListBuilder) Build(layoutRoot *LayoutBox, renderRoot *RenderNode) *DisplayList {
	displayList := NewDisplayList()
	
	if layoutRoot == nil || renderRoot == nil {
		return displayList
	}
	
	// Build a map of render nodes by ID for quick lookup
	renderMap := dlb.buildRenderMap(renderRoot)
	
	// Walk the layout tree and generate paint commands
	dlb.buildRecursive(layoutRoot, renderMap, displayList)
	
	return displayList
}

// buildRenderMap builds a map of render nodes indexed by their ID
func (dlb *DisplayListBuilder) buildRenderMap(root *RenderNode) map[int64]*RenderNode {
	nodeMap := make(map[int64]*RenderNode)
	dlb.buildRenderMapRecursive(root, nodeMap)
	return nodeMap
}

// buildRenderMapRecursive recursively builds the render node map
func (dlb *DisplayListBuilder) buildRenderMapRecursive(node *RenderNode, nodeMap map[int64]*RenderNode) {
	if node == nil {
		return
	}
	
	nodeMap[node.ID] = node
	
	for _, child := range node.Children {
		dlb.buildRenderMapRecursive(child, nodeMap)
	}
}

// buildRecursive recursively builds paint commands for a layout box
func (dlb *DisplayListBuilder) buildRecursive(layoutBox *LayoutBox, renderMap map[int64]*RenderNode, displayList *DisplayList) {
	if layoutBox == nil {
		return
	}
	
	// Get the corresponding render node
	renderNode, exists := renderMap[layoutBox.NodeID]
	if !exists {
		return
	}
	
	// Generate paint command based on node type
	if renderNode.Type == NodeTypeText {
		dlb.addTextCommand(layoutBox, renderNode, displayList)
	} else if renderNode.Type == NodeTypeElement {
		dlb.addElementCommand(layoutBox, renderNode, displayList)
	}
	
	// Process children
	for _, child := range layoutBox.Children {
		dlb.buildRecursive(child, renderMap, displayList)
	}
}

// addTextCommand adds a text paint command
func (dlb *DisplayListBuilder) addTextCommand(layoutBox *LayoutBox, renderNode *RenderNode, displayList *DisplayList) {
	text := renderNode.Text
	if text == "" {
		return
	}
	
	// Get text style from node hierarchy
	style := dlb.fontMetrics.GetTextStyleFromNode(renderNode)
	
	// Get font size from parent
	fontSize := dlb.defaultFontSize
	if renderNode.Parent != nil {
		fontSize = dlb.fontMetrics.GetFontSize(renderNode.Parent.TagName)
	}
	
	cmd := &PaintCommand{
		Type:     PaintText,
		NodeID:   layoutBox.NodeID,
		Box:      layoutBox.Box,
		Text:     text,
		FontSize: fontSize,
		Bold:     style.Bold,
		Italic:   style.Italic,
	}
	
	displayList.AddCommand(cmd)
}

// addElementCommand adds paint commands for an element
func (dlb *DisplayListBuilder) addElementCommand(layoutBox *LayoutBox, renderNode *RenderNode, displayList *DisplayList) {
	// For image elements, add a rectangle placeholder and text
	if renderNode.TagName == "img" {
		// Add background rectangle
		cmd := &PaintCommand{
			Type:        PaintRect,
			NodeID:      layoutBox.NodeID,
			Box:         layoutBox.Box,
			FillColor:   color.RGBA{R: 200, G: 200, B: 200, A: 255},
			StrokeColor: color.RGBA{R: 150, G: 150, B: 150, A: 255},
			StrokeWidth: 1.0,
		}
		displayList.AddCommand(cmd)
		
		// Add image info text if available
		src, _ := renderNode.GetAttribute("src")
		alt, _ := renderNode.GetAttribute("alt")
		
		if src != "" || alt != "" {
			textCmd := &PaintCommand{
				Type:     PaintImage,
				NodeID:   layoutBox.NodeID,
				Box:      layoutBox.Box,
				ImageSrc: src,
				ImageAlt: alt,
			}
			displayList.AddCommand(textCmd)
		}
	}
	
	// For other elements, we primarily rely on their children for rendering
	// but we could add background colors, borders, etc. here in the future
}
