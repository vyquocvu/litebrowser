package js

import (
	"strings"
	"testing"
)

func TestNewRuntime(t *testing.T) {
	runtime := NewRuntime()
	if runtime == nil {
		t.Fatal("NewRuntime() returned nil")
	}
	if runtime.vm == nil {
		t.Fatal("Runtime vm is nil")
	}
}

func TestConsoleLog(t *testing.T) {
	runtime := NewRuntime()
	
	_, err := runtime.RunScript(`console.log("test message");`)
	if err != nil {
		t.Errorf("console.log failed: %v", err)
	}
}

func TestDocumentGetElementByID(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body><div id="test">Test Content</div></body></html>`
	runtime.SetHTMLContent(html)
	
	// Test getting non-existent element
	val, err := runtime.RunScript(`document.getElementById("nonexistent");`)
	if err != nil {
		t.Errorf("getElementById failed: %v", err)
	}
	if !val.ToBoolean() {
		// Should be null/falsy for non-existent element
		t.Log("Correctly returned null for non-existent element")
	}
	
	// Test getting existing element
	val, err = runtime.RunScript(`
		var elem = document.getElementById("test");
		elem ? elem.textContent : null;
	`)
	if err != nil {
		t.Errorf("getElementById failed: %v", err)
	}
	if val.Export() != nil {
		result := val.String()
		if !strings.Contains(result, "Test Content") {
			t.Errorf("Expected textContent to contain 'Test Content', got %v", result)
		}
	}
}

func TestSetHTMLContent(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>Test</body></html>`
	runtime.SetHTMLContent(html)
	
	if runtime.htmlCache != html {
		t.Errorf("SetHTMLContent() did not set htmlCache correctly")
	}
}

func TestRunScript(t *testing.T) {
	runtime := NewRuntime()
	
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:    "valid script",
			script:  `var x = 1 + 1;`,
			wantErr: false,
		},
		{
			name:    "console.log",
			script:  `console.log("test");`,
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			script:  `var x = ;`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runtime.RunScript(tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetElementsByClassName(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>
		<div class="item">Item 1</div>
		<p class="item">Item 2</p>
		<span class="other">Other</span>
	</body></html>`
	runtime.SetHTMLContent(html)
	
	val, err := runtime.RunScript(`
		var elements = document.getElementsByClassName("item");
		elements.length;
	`)
	if err != nil {
		t.Errorf("getElementsByClassName failed: %v", err)
	}
	if val.ToInteger() != 2 {
		t.Errorf("Expected 2 elements, got %d", val.ToInteger())
	}
}

func TestGetElementsByTagName(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>
		<div>Div 1</div>
		<div>Div 2</div>
		<p>Paragraph</p>
	</body></html>`
	runtime.SetHTMLContent(html)
	
	val, err := runtime.RunScript(`
		var elements = document.getElementsByTagName("div");
		elements.length;
	`)
	if err != nil {
		t.Errorf("getElementsByTagName failed: %v", err)
	}
	if val.ToInteger() != 2 {
		t.Errorf("Expected 2 elements, got %d", val.ToInteger())
	}
}

func TestQuerySelector(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>
		<div id="main" class="container">Main Content</div>
		<p class="text">Paragraph</p>
	</body></html>`
	runtime.SetHTMLContent(html)
	
	tests := []struct {
		name     string
		script   string
		wantNull bool
	}{
		{
			name:     "ID selector",
			script:   `document.querySelector("#main")`,
			wantNull: false,
		},
		{
			name:     "class selector",
			script:   `document.querySelector(".text")`,
			wantNull: false,
		},
		{
			name:     "tag selector",
			script:   `document.querySelector("p")`,
			wantNull: false,
		},
		{
			name:     "non-matching selector",
			script:   `document.querySelector("#nonexistent")`,
			wantNull: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := runtime.RunScript(tt.script)
			if err != nil {
				t.Errorf("querySelector failed: %v", err)
			}
			isNull := val == nil || val.String() == "null"
			if isNull != tt.wantNull {
				t.Errorf("querySelector returned null=%v, want null=%v", isNull, tt.wantNull)
			}
		})
	}
}

func TestQuerySelectorAll(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>
		<div class="item">Item 1</div>
		<div class="item">Item 2</div>
		<p class="item">Item 3</p>
	</body></html>`
	runtime.SetHTMLContent(html)
	
	val, err := runtime.RunScript(`
		var elements = document.querySelectorAll(".item");
		elements.length;
	`)
	if err != nil {
		t.Errorf("querySelectorAll failed: %v", err)
	}
	if val.ToInteger() != 3 {
		t.Errorf("Expected 3 elements, got %d", val.ToInteger())
	}
}

func TestCreateElement(t *testing.T) {
	runtime := NewRuntime()
	
	val, err := runtime.RunScript(`
		var div = document.createElement("div");
		div.tagName;
	`)
	if err != nil {
		t.Errorf("createElement failed: %v", err)
	}
	if val.String() != "div" {
		t.Errorf("Expected tagName 'div', got %s", val.String())
	}
}

func TestAppendChild(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body><div id="parent">Parent</div></body></html>`
	runtime.SetHTMLContent(html)
	
	val, err := runtime.RunScript(`
		var parent = document.getElementById("parent");
		var child = document.createElement("span");
		child.textContent = "Child";
		parent.appendChild(child);
		parent.children.length;
	`)
	if err != nil {
		t.Errorf("appendChild failed: %v", err)
	}
	if val.ToInteger() != 1 {
		t.Errorf("Expected 1 child, got %d", val.ToInteger())
	}
}

func TestRemoveChild(t *testing.T) {
	runtime := NewRuntime()
	
	val, err := runtime.RunScript(`
		var parent = document.createElement("div");
		var child1 = document.createElement("span");
		var child2 = document.createElement("p");
		parent.appendChild(child1);
		parent.appendChild(child2);
		parent.removeChild(child1);
		parent.children.length;
	`)
	if err != nil {
		t.Errorf("removeChild failed: %v", err)
	}
	if val.ToInteger() != 1 {
		t.Errorf("Expected 1 child after removal, got %d", val.ToInteger())
	}
}

func TestReplaceChild(t *testing.T) {
	runtime := NewRuntime()
	
	val, err := runtime.RunScript(`
		var parent = document.createElement("div");
		var oldChild = document.createElement("span");
		var newChild = document.createElement("p");
		oldChild.textContent = "Old";
		newChild.textContent = "New";
		parent.appendChild(oldChild);
		parent.replaceChild(newChild, oldChild);
		parent.children[0].textContent;
	`)
	if err != nil {
		t.Errorf("replaceChild failed: %v", err)
	}
	if val.String() != "New" {
		t.Errorf("Expected 'New', got %s", val.String())
	}
}

func TestInsertBefore(t *testing.T) {
	runtime := NewRuntime()
	
	val, err := runtime.RunScript(`
		var parent = document.createElement("div");
		var child1 = document.createElement("span");
		var child2 = document.createElement("p");
		child1.textContent = "First";
		child2.textContent = "Second";
		parent.appendChild(child2);
		parent.insertBefore(child1, child2);
		parent.children[0].textContent;
	`)
	if err != nil {
		t.Errorf("insertBefore failed: %v", err)
	}
	if val.String() != "First" {
		t.Errorf("Expected 'First' at index 0, got %s", val.String())
	}
}

func TestAddEventListener(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body><button id="btn">Click</button></body></html>`
	runtime.SetHTMLContent(html)
	
	_, err := runtime.RunScript(`
		var btn = document.getElementById("btn");
		var clicked = false;
		btn.addEventListener("click", function() {
			clicked = true;
		});
	`)
	if err != nil {
		t.Errorf("addEventListener failed: %v", err)
	}
}

func TestRemoveEventListener(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body><button id="btn">Click</button></body></html>`
	runtime.SetHTMLContent(html)
	
	_, err := runtime.RunScript(`
		var btn = document.getElementById("btn");
		var handler = function() {
			console.log("clicked");
		};
		btn.addEventListener("click", handler);
		btn.removeEventListener("click", handler);
	`)
	if err != nil {
		t.Errorf("removeEventListener failed: %v", err)
	}
}

func TestElementProperties(t *testing.T) {
	runtime := NewRuntime()
	
	html := `<html><body>
		<div id="test" class="container active">Test Content</div>
	</body></html>`
	runtime.SetHTMLContent(html)
	
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name:   "tagName",
			script: `document.querySelector("#test").tagName`,
			want:   "div",
		},
		{
			name:   "id",
			script: `document.querySelector("#test").id`,
			want:   "test",
		},
		{
			name:   "textContent",
			script: `document.querySelector("#test").textContent`,
			want:   "Test Content",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := runtime.RunScript(tt.script)
			if err != nil {
				t.Errorf("Failed to get property: %v", err)
			}
			got := val.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("Expected to contain %q, got %q", tt.want, got)
			}
		})
	}
}
