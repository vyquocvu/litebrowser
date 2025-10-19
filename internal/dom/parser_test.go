package dom

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}
}

func TestParseBodyText(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		html     string
		wantText string
		wantErr  bool
	}{
		{
			name: "simple body",
			html: `<html><body>Hello World</body></html>`,
			wantText: "Hello World",
			wantErr: false,
		},
		{
			name: "body with nested elements",
			html: `<html><body><h1>Title</h1><p>Paragraph</p></body></html>`,
			wantText: "Title Paragraph",
			wantErr: false,
		},
		{
			name: "empty body",
			html: `<html><body></body></html>`,
			wantText: "",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseBodyText(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantText {
				t.Errorf("ParseBodyText() = %v, want %v", got, tt.wantText)
			}
		})
	}
}

func TestGetElementByID(t *testing.T) {
	parser := NewParser()
	
	html := `<html><body><div id="test">Test Content</div><p id="para">Paragraph</p></body></html>`
	
	tests := []struct {
		name     string
		id       string
		wantText string
		wantErr  bool
	}{
		{
			name: "existing id",
			id: "test",
			wantText: "Test Content",
			wantErr: false,
		},
		{
			name: "another existing id",
			id: "para",
			wantText: "Paragraph",
			wantErr: false,
		},
		{
			name: "non-existing id",
			id: "nonexistent",
			wantText: "",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.GetElementByID(html, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetElementByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.wantText) && tt.wantText != "" {
				t.Errorf("GetElementByID() = %v, want to contain %v", got, tt.wantText)
			}
		})
	}
}
