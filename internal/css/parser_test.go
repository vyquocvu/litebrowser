package css

import (
	"testing"
)

func TestParser(t *testing.T) {
	css := `
		h1 {
			display: block;
			font-size: 32px;
		}
	`
	p := NewParser(css)
	stylesheet, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}
	if len(stylesheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(stylesheet.Rules))
	}
	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 1 {
		t.Fatalf("expected 1 selector, got %d", len(rule.Selectors))
	}
	selector := rule.Selectors[0]
	if selector.TagName != "h1" {
		t.Errorf("expected tag name 'h1', got '%s'", selector.TagName)
	}
	if len(rule.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(rule.Declarations))
	}
	decl1 := rule.Declarations[0]
	if decl1.Property != "display" || decl1.Value != "block" {
		t.Errorf("unexpected declaration: %s: %s", decl1.Property, decl1.Value)
	}
	decl2 := rule.Declarations[1]
	if decl2.Property != "font-size" || decl2.Value != "32px" {
		t.Errorf("unexpected declaration: %s: %s", decl2.Property, decl2.Value)
	}
}

func TestParserCombinedSelector(t *testing.T) {
	css := `
		h1.title {
			color: blue;
		}
	`
	p := NewParser(css)
	stylesheet, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}
	if len(stylesheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(stylesheet.Rules))
	}
	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 1 {
		t.Fatalf("expected 1 selector, got %d", len(rule.Selectors))
	}
	selector := rule.Selectors[0]
	if selector.TagName != "h1" {
		t.Errorf("expected tag name 'h1', got '%s'", selector.TagName)
	}
	if len(selector.Classes) != 1 || selector.Classes[0] != "title" {
		t.Errorf("expected class 'title', got %v", selector.Classes)
	}
	if len(rule.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(rule.Declarations))
	}
	decl := rule.Declarations[0]
	if decl.Property != "color" || decl.Value != "blue" {
		t.Errorf("unexpected declaration: %s: %s", decl.Property, decl.Value)
	}
}
