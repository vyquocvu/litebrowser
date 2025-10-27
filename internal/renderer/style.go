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
		for _, selector := range rule.Selectors {
			if sm.matches(selector, node) {
				for _, decl := range rule.Declarations {
					sm.applyDeclaration(node, decl)
				}
			}
		}
	}
}

func (sm *StyleManager) matches(selector css.Selector, node *RenderNode) bool {
	if selector.TagName != "" && selector.TagName != node.TagName {
		return false
	}
	if selector.ID != "" {
		id, ok := node.GetAttribute("id")
		if !ok || id != selector.ID {
			return false
		}
	}
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
	if selector.PseudoClass != "" {
		switch selector.PseudoClass {
		case "link", "visited":
			if node.TagName != "a" {
				return false
			}
		default:
			return false // Unsupported pseudo-class
		}
	}
	return true
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
	case "background":
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
	if strings.HasPrefix(value, "#") {
		return parseHexColor(value)
	}
	// Add support for other color formats like rgb() later
	return color.Black, fmt.Errorf("unsupported color format")
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
