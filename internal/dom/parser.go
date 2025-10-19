package dom

import (
	"strings"

	"golang.org/x/net/html"
)

// Parser handles HTML parsing
type Parser struct{}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{}
}

// ParseBodyText extracts text content from the body element
func (p *Parser) ParseBodyText(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var bodyText strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			p.getTextFromNode(n, &bodyText)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)
	return strings.TrimSpace(bodyText.String()), nil
}

// getTextFromNode extracts text from a node and its children
func (p *Parser) getTextFromNode(n *html.Node, builder *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(text)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.getTextFromNode(c, builder)
	}
}

// GetElementByID searches for an element by ID (basic implementation)
func (p *Parser) GetElementByID(htmlContent, id string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var result string
	var findElement func(*html.Node)
	findElement = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == id {
					var textBuilder strings.Builder
					p.getTextFromNode(n, &textBuilder)
					result = textBuilder.String()
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result == "" {
				findElement(c)
			}
		}
	}

	findElement(doc)
	return result, nil
}
