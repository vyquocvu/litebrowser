package renderer

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

func TestFormElementRendering(t *testing.T) {
	htmlContent := `
		<html>
			<body>
				<input placeholder="Enter text" />
				<button>Click me</button>
				<textarea placeholder="Enter more text"></textarea>
			</body>
		</html>
	`
	r := NewRenderer(800, 600)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}
	renderTree := BuildRenderTree(findBodyNode(doc))
	obj := r.canvasRenderer.Render(renderTree)

	// The top-level object is a container, let's inspect its children
	topContainer, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("Expected a container, but got %T", obj)
	}

	if len(topContainer.Objects) != 3 {
		t.Fatalf("Expected 3 objects, got %d", len(topContainer.Objects))
	}

	if _, ok := topContainer.Objects[0].(*widget.Entry); !ok {
		t.Errorf("Expected first object to be an Entry, but it was not")
	}
	if _, ok := topContainer.Objects[1].(*widget.Button); !ok {
		t.Errorf("Expected second object to be a Button, but it was not")
	}
	if _, ok := topContainer.Objects[2].(*widget.Entry); !ok {
		t.Errorf("Expected third object to be a MultiLineEntry, but it was not")
	}
}

func TestTableElementRendering(t *testing.T) {
	htmlContent := `
		<html>
			<body>
				<table>
					<tr>
						<td>Cell 1</td>
						<td>Cell 2</td>
					</tr>
					<tr>
						<td>Cell 3</td>
						<td>Cell 4</td>
					</tr>
				</table>
			</body>
		</html>
	`
	r := NewRenderer(800, 600)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("html.Parse failed: %v", err)
	}
	renderTree := BuildRenderTree(findBodyNode(doc))
	obj := r.canvasRenderer.Render(renderTree)

	topContainer, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("Expected a container, but got %T", obj)
	}

	if len(topContainer.Objects) != 1 {
		t.Fatalf("Expected 1 object, got %d", len(topContainer.Objects))
	}

	form, ok := topContainer.Objects[0].(*widget.Form)
	if !ok {
		t.Fatalf("Expected a Form widget, but got something else")
	}

	if len(form.Items) != 4 {
		t.Errorf("Expected 4 form items, but got %d", len(form.Items))
	}
}
