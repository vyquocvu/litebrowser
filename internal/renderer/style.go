package renderer

import (
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

	sm.applyMatchingRules(node)

	for _, child := range node.Children {
		sm.ApplyStyles(child)
	}
}

func (sm *StyleManager) applyMatchingRules(node *RenderNode) {
	if node.ComputedStyle == nil {
		node.ComputedStyle = &Style{}
	}

	for _, rule := range sm.stylesheet.Rules {
		for _, selector := range rule.Selectors {
			if sm.matches(selector, node) {
				for _, decl := range rule.Declarations {
					sm.applyDeclaration(node.ComputedStyle, decl)
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
	return true
}

func (sm *StyleManager) applyDeclaration(style *Style, decl css.Declaration) {
	switch decl.Property {
	case "display":
		style.Display = decl.Value
	case "font-size":
		if val, err := parseFontSize(decl.Value); err == nil {
			style.FontSize = val
		}
	case "font-weight":
		style.FontWeight = decl.Value
	case "color":
		style.Color = decl.Value
	}
}

func parseFontSize(value string) (float32, error) {
	if strings.HasSuffix(value, "px") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(value, "px"), 32)
		if err != nil {
			return 0, err
		}
		return float32(val), nil
	}
	// For now, we only support px. Other units can be added later.
	return 0, &strconv.NumError{}
}
