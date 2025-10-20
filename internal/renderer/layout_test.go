package renderer

import (
	"testing"
)

func TestNewLayoutEngine(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	if le == nil {
		t.Fatal("NewLayoutEngine returned nil")
	}
	if le.canvasWidth != 800 {
		t.Errorf("Expected width 800, got %f", le.canvasWidth)
	}
	if le.canvasHeight != 600 {
		t.Errorf("Expected height 600, got %f", le.canvasHeight)
	}
}

func TestLayoutTextNode(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	node := NewRenderNode(NodeTypeText)
	node.Text = "Hello World"
	
	le.Layout(node)
	
	if node.Box.X != 0 {
		t.Errorf("Expected X=0, got %f", node.Box.X)
	}
	if node.Box.Y != 0 {
		t.Errorf("Expected Y=0, got %f", node.Box.Y)
	}
	if node.Box.Width != 800 {
		t.Errorf("Expected Width=800, got %f", node.Box.Width)
	}
	if node.Box.Height <= 0 {
		t.Error("Expected Height > 0")
	}
}

func TestLayoutElementNode(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	node := NewRenderNode(NodeTypeElement)
	node.TagName = "div"
	
	child := NewRenderNode(NodeTypeText)
	child.Text = "Content"
	node.AddChild(child)
	
	le.Layout(node)
	
	if node.Box.X != 0 {
		t.Errorf("Expected X=0, got %f", node.Box.X)
	}
	if node.Box.Width != 800 {
		t.Errorf("Expected Width=800, got %f", node.Box.Width)
	}
	if node.Box.Height <= 0 {
		t.Error("Expected Height > 0")
	}
	
	// Check child was laid out
	if child.Box.Width <= 0 {
		t.Error("Child box width not set")
	}
}

func TestLayoutHeading(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	
	tests := []struct {
		tagName      string
		expectedSize float32
	}{
		{"h1", le.defaultFontSize * 2.0},
		{"h2", le.defaultFontSize * 1.5},
		{"h3", le.defaultFontSize * 1.17},
		{"p", le.defaultFontSize},
	}
	
	for _, tt := range tests {
		t.Run(tt.tagName, func(t *testing.T) {
			size := le.getFontSize(tt.tagName)
			if size != tt.expectedSize {
				t.Errorf("Expected font size %f for %s, got %f", tt.expectedSize, tt.tagName, size)
			}
		})
	}
}

func TestLayoutVerticalSpacing(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	
	tests := []struct {
		tagName string
		hasSpacing bool
	}{
		{"h1", true},
		{"h2", true},
		{"p", true},
		{"ul", true},
		{"li", true},
		{"span", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.tagName, func(t *testing.T) {
			spacing := le.getVerticalSpacing(tt.tagName)
			if tt.hasSpacing && spacing <= 0 {
				t.Errorf("Expected spacing > 0 for %s, got %f", tt.tagName, spacing)
			}
			if !tt.hasSpacing && spacing != 0 {
				t.Errorf("Expected spacing = 0 for %s, got %f", tt.tagName, spacing)
			}
		})
	}
}

func TestLayoutMultipleChildren(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	parent := NewRenderNode(NodeTypeElement)
	parent.TagName = "div"
	
	// Add multiple children
	for i := 0; i < 3; i++ {
		child := NewRenderNode(NodeTypeElement)
		child.TagName = "p"
		text := NewRenderNode(NodeTypeText)
		text.Text = "Paragraph text"
		child.AddChild(text)
		parent.AddChild(child)
	}
	
	le.Layout(parent)
	
	// Verify children are stacked vertically
	prevY := float32(0)
	for i, child := range parent.Children {
		if i > 0 && child.Box.Y <= prevY {
			t.Errorf("Child %d should be positioned below previous child", i)
		}
		prevY = child.Box.Y + child.Box.Height
	}
}

func TestLayoutNestedElements(t *testing.T) {
	le := NewLayoutEngine(800, 600)
	
	// Create nested structure: div > p > text
	div := NewRenderNode(NodeTypeElement)
	div.TagName = "div"
	
	p := NewRenderNode(NodeTypeElement)
	p.TagName = "p"
	
	text := NewRenderNode(NodeTypeText)
	text.Text = "Nested text"
	
	p.AddChild(text)
	div.AddChild(p)
	
	le.Layout(div)
	
	// Verify all nodes have been laid out
	if div.Box.Width == 0 {
		t.Error("Div box width not set")
	}
	if p.Box.Width == 0 {
		t.Error("P box width not set")
	}
	if text.Box.Width == 0 {
		t.Error("Text box width not set")
	}
}
