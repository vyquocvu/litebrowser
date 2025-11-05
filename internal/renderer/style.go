package renderer

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/vyquocvu/goosie/internal/css"
)

// StyleManager applies styles from a stylesheet to a render tree.
type StyleManager struct {
	stylesheet *css.StyleSheet
}

// NewStyleManager creates a new StyleManager.
func NewStyleManager(stylesheet *css.StyleSheet) *StyleManager {
	return &StyleManager{stylesheet: stylesheet}
}

// ApplyStyles applies the styles to the given render tree.
func (sm *StyleManager) ApplyStyles(node *RenderNode) {
	if node == nil {
		return
	}

	if node.ComputedStyle == nil {
		node.ComputedStyle = &Style{}
	}

	// Inherit styles from parent
	if node.Parent != nil && node.Parent.ComputedStyle != nil {
		node.ComputedStyle.Color = node.Parent.ComputedStyle.Color
		node.ComputedStyle.FontSize = node.Parent.ComputedStyle.FontSize
	}

	sm.applyMatchingRules(node)

	for _, child := range node.Children {
		sm.ApplyStyles(child)
	}
}

func (sm *StyleManager) applyMatchingRules(node *RenderNode) {
	for _, rule := range sm.stylesheet.Rules {
		for _, selectorSeq := range rule.Selectors {
			if sm.matchesSequence(selectorSeq, node) {
				for _, decl := range rule.Declarations {
					sm.applyDeclaration(node, decl)
				}
			}
		}
	}
}

// matchesSequence checks if a selector sequence matches a node
func (sm *StyleManager) matchesSequence(seq css.SelectorSequence, node *RenderNode) bool {
	// For left-to-right parsing: div > p
	// We need to match from right to left: check if node matches 'p', then check if parent matches 'div'
	return sm.matchesFromRight(&seq, node)
}

// matchesFromRight recursively matches selectors from right to left
func (sm *StyleManager) matchesFromRight(seq *css.SelectorSequence, node *RenderNode) bool {
	// Find the rightmost selector in the chain
	if seq.Next == nil {
		// This is the rightmost selector, match it against the node
		return sm.matchesSimple(seq.Simple, node)
	}
	
	// This is not the rightmost, so we need to match the rightmost first
	// and then check if this one matches the appropriate ancestor/sibling
	return sm.matchesWithCombinatorLeftToRight(seq, node)
}

// matchesWithCombinatorLeftToRight handles combinator matching for left-to-right sequences
func (sm *StyleManager) matchesWithCombinatorLeftToRight(seq *css.SelectorSequence, node *RenderNode) bool {
	// seq = A (combinator) B
	// We need to check if B matches the node, then verify A matches the related element
	
	// First, recursively match the right side
	if !sm.matchesFromRight(seq.Next, node) {
		return false
	}
	
	// Now check if the left side (seq.Simple) matches according to the combinator
	switch seq.Combinator {
	case " ": // Descendant combinator: A B means B is descendant of A
		return sm.hasMatchingAncestor(seq.Simple, node)
	case ">": // Child combinator: A > B means B is direct child of A
		if node.Parent != nil {
			return sm.matchesSimple(seq.Simple, node.Parent)
		}
		return false
	case "+": // Adjacent sibling: A + B means B immediately follows A
		sibling := sm.getPreviousSibling(node)
		if sibling != nil {
			return sm.matchesSimple(seq.Simple, sibling)
		}
		return false
	case "~": // General sibling: A ~ B means B is preceded by A
		return sm.hasMatchingPreviousSibling(seq.Simple, node)
	}
	
	return false
}

// hasMatchingAncestor checks if any ancestor matches the selector
func (sm *StyleManager) hasMatchingAncestor(selector css.SimpleSelector, node *RenderNode) bool {
	current := node.Parent
	for current != nil {
		if sm.matchesSimple(selector, current) {
			return true
		}
		current = current.Parent
	}
	return false
}

// hasMatchingPreviousSibling checks if any previous sibling matches the selector
func (sm *StyleManager) hasMatchingPreviousSibling(selector css.SimpleSelector, node *RenderNode) bool {
	if node.Parent == nil {
		return false
	}
	
	// Find node's index in parent's children
	nodeIndex := -1
	for i, child := range node.Parent.Children {
		if child == node {
			nodeIndex = i
			break
		}
	}
	
	if nodeIndex == -1 {
		return false
	}
	
	// Check all previous siblings
	for i := nodeIndex - 1; i >= 0; i-- {
		if sm.matchesSimple(selector, node.Parent.Children[i]) {
			return true
		}
	}
	
	return false
}

// getPreviousSibling returns the previous sibling of a node
func (sm *StyleManager) getPreviousSibling(node *RenderNode) *RenderNode {
	if node.Parent == nil {
		return nil
	}
	
	for i, child := range node.Parent.Children {
		if child == node && i > 0 {
			return node.Parent.Children[i-1]
		}
	}
	
	return nil
}

// matchesSimple checks if a simple selector matches a node
func (sm *StyleManager) matchesSimple(selector css.SimpleSelector, node *RenderNode) bool {
	// Universal selector matches everything
	if selector.Universal && selector.ID == "" && len(selector.Classes) == 0 && 
		len(selector.PseudoClasses) == 0 && len(selector.Attributes) == 0 {
		return true
	}
	
	// Check tag name
	if selector.TagName != "" && selector.TagName != node.TagName {
		return false
	}
	
	// Check ID
	if selector.ID != "" {
		id, ok := node.GetAttribute("id")
		if !ok || id != selector.ID {
			return false
		}
	}
	
	// Check classes
	if len(selector.Classes) > 0 {
		classAttr, ok := node.GetAttribute("class")
		if !ok {
			return false
		}
		classes := strings.Fields(classAttr)
		for _, selClass := range selector.Classes {
			found := false
			for _, nodeClass := range classes {
				if selClass == nodeClass {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	// Check pseudo-classes
	for _, pseudoClass := range selector.PseudoClasses {
		if !sm.matchesPseudoClass(pseudoClass, node) {
			return false
		}
	}
	
	// Check attributes
	for _, attr := range selector.Attributes {
		if !sm.matchesAttribute(attr, node) {
			return false
		}
	}
	
	return true
}

// matchesPseudoClass checks if a pseudo-class matches a node
func (sm *StyleManager) matchesPseudoClass(pseudoClass string, node *RenderNode) bool {
	// Handle functional pseudo-classes
	if strings.HasPrefix(pseudoClass, "nth-child(") {
		// For now, just return true for basic support
		return true
	}
	
	switch pseudoClass {
	case "link", "visited":
		return node.TagName == "a"
	case "hover", "focus", "active":
		// These require state tracking, not implemented yet
		return false
	case "first-child":
		if node.Parent == nil || len(node.Parent.Children) == 0 {
			return false
		}
		return node.Parent.Children[0] == node
	case "last-child":
		if node.Parent == nil || len(node.Parent.Children) == 0 {
			return false
		}
		return node.Parent.Children[len(node.Parent.Children)-1] == node
	default:
		return false
	}
}

// matchesAttribute checks if an attribute selector matches a node
func (sm *StyleManager) matchesAttribute(attr css.AttributeSelector, node *RenderNode) bool {
	value, ok := node.GetAttribute(attr.Name)
	if !ok {
		return false
	}
	
	if attr.Operator == "" {
		// Just checking for attribute presence
		return true
	}
	
	switch attr.Operator {
	case "=":
		return value == attr.Value
	case "~=":
		// Word match
		words := strings.Fields(value)
		for _, word := range words {
			if word == attr.Value {
				return true
			}
		}
		return false
	case "|=":
		// Exact or prefix with hyphen
		return value == attr.Value || strings.HasPrefix(value, attr.Value+"-")
	case "^=":
		// Starts with
		return strings.HasPrefix(value, attr.Value)
	case "$=":
		// Ends with
		return strings.HasSuffix(value, attr.Value)
	case "*=":
		// Contains
		return strings.Contains(value, attr.Value)
	}
	
	return false
}

// Legacy function for backward compatibility with old Selector type
func (sm *StyleManager) matches(selector css.SimpleSelector, node *RenderNode) bool {
	return sm.matchesSimple(selector, node)
}

// colorNameToHex maps common color names to their hex values.
var colorNameToHex = map[string]string{
	"black":   "#000000",
	"white":   "#ffffff",
	"red":     "#ff0000",
	"green":   "#008000",
	"blue":    "#0000ff",
	"yellow":  "#ffff00",
	"cyan":    "#00ffff",
	"magenta": "#ff00ff",
	"silver":  "#c0c0c0",
	"gray":    "#808080",
	"maroon":  "#800000",
	"olive":   "#808000",
	"purple":  "#800080",
	"teal":    "#008080",
	"navy":    "#000080",
}

func (sm *StyleManager) applyDeclaration(node *RenderNode, decl css.Declaration) {
	style := node.ComputedStyle
	switch decl.Property {
	case "display":
		style.Display = decl.Value
	case "font-size":
		parentFontSize := float32(16.0) // Default font size
		if node.Parent != nil && node.Parent.ComputedStyle != nil && node.Parent.ComputedStyle.FontSize > 0 {
			parentFontSize = node.Parent.ComputedStyle.FontSize
		}
		if val, err := parseFontSize(decl.Value, parentFontSize); err == nil {
			style.FontSize = val
		}
	case "font-weight":
		style.FontWeight = decl.Value
	case "color":
		if val, err := parseColor(decl.Value); err == nil {
			style.Color = val
		}
	case "background-color":
		if val, err := parseColor(decl.Value); err == nil {
			style.BackgroundColor = val
		}
	case "width":
		style.Width = decl.Value
	case "margin":
		style.Margin = decl.Value
	case "font-family":
		style.FontFamily = decl.Value
	case "opacity":
		if val, err := strconv.ParseFloat(decl.Value, 32); err == nil {
			style.Opacity = float32(val)
		}
	}
}

func parseFontSize(value string, parentFontSize float32) (float32, error) {
	if strings.HasSuffix(value, "px") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(value, "px"), 32)
		if err != nil {
			return 0, err
		}
		return float32(val), nil
	}
	if strings.HasSuffix(value, "em") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(value, "em"), 32)
		if err != nil {
			return 0, err
		}
		return float32(val) * parentFontSize, nil
	}
	return 0, fmt.Errorf("unsupported font size unit")
}

func parseColor(value string) (color.Color, error) {
	lowerValue := strings.ToLower(value)
	if hex, ok := colorNameToHex[lowerValue]; ok {
		return parseHexColor(hex)
	}
	if strings.HasPrefix(lowerValue, "#") {
		return parseHexColor(lowerValue)
	}
	// Add support for other color formats like rgb() later
	return color.Black, fmt.Errorf("unsupported color format: %s", value)
}

func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color length")
	}
	rgb, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return nil, err
	}
	return color.RGBA{
		R: uint8(rgb >> 16),
		G: uint8(rgb >> 8),
		B: uint8(rgb),
		A: 255,
	}, nil
}
