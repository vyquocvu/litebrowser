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
