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

// ParseBodyHTML extracts HTML content from the body element and converts to markdown-like format
func (p *Parser) ParseBodyHTML(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var markdown strings.Builder
	var extractHTML func(*html.Node)
	extractHTML = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			p.convertToMarkdown(n, &markdown)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractHTML(c)
		}
	}

	extractHTML(doc)
	return strings.TrimSpace(markdown.String()), nil
}

// convertToMarkdown converts HTML nodes to markdown-like format
func (p *Parser) convertToMarkdown(n *html.Node, builder *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			builder.WriteString(text)
		}
	case html.ElementNode:
		switch n.Data {
		case "h1":
			builder.WriteString("\n# ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("\n\n")
		case "h2":
			builder.WriteString("\n## ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("\n\n")
		case "h3":
			builder.WriteString("\n### ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("\n\n")
		case "p":
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("\n\n")
		case "a":
			href := ""
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			builder.WriteString("[")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("]")
			if href != "" {
				builder.WriteString("(")
				builder.WriteString(href)
				builder.WriteString(")")
			}
		case "strong", "b":
			builder.WriteString("**")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("**")
		case "em", "i":
			builder.WriteString("*")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("*")
		case "br":
			builder.WriteString("\n")
		case "div":
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
			builder.WriteString("\n")
		default:
			// For other elements, just process children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				p.convertToMarkdown(c, builder)
			}
		}
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			p.convertToMarkdown(c, builder)
		}
	}
}
