package css

import (
	"fmt"
	"strings"
)

// Parser processes CSS text and builds a StyleSheet.
type Parser struct {
	input string
	pos   int
}

// NewParser creates a new Parser.
func NewParser(input string) *Parser {
	return &Parser{input: input}
}

// Parse parses the CSS input and returns a StyleSheet.
func (p *Parser) Parse() (*StyleSheet, error) {
	stylesheet := &StyleSheet{}
	p.consumeWhitespace()
	for p.pos < len(p.input) {
		selectors, err := p.parseSelectors()
		if err != nil {
			return nil, err
		}
		p.consumeWhitespace()
		if !p.consumeChar('{') {
			return nil, fmt.Errorf("expected '{'")
		}
		p.consumeWhitespace()
		declarations := p.parseDeclarations()
		p.consumeWhitespace()
		if !p.consumeChar('}') {
			return nil, fmt.Errorf("expected '}'")
		}
		stylesheet.Rules = append(stylesheet.Rules, Rule{Selectors: selectors, Declarations: declarations})
		p.consumeWhitespace()
	}
	return stylesheet, nil
}

func (p *Parser) parseSelectors() ([]Selector, error) {
	var selectors []Selector
	for {
		selector := Selector{}
		p.consumeWhitespace()
		// First, try to parse a tag name.
		if isIdentifierChar(p.peek()) {
			selector.TagName = p.consumeIdentifier()
		}
		// Then, parse any classes and IDs.
		for {
			if p.peek() == '#' {
				p.consumeChar('#')
				selector.ID = p.consumeIdentifier()
			} else if p.peek() == '.' {
				p.consumeChar('.')
				selector.Classes = append(selector.Classes, p.consumeIdentifier())
			} else {
				break
			}
		}

		selectors = append(selectors, selector)

		p.consumeWhitespace()
		if p.peek() == ',' {
			p.consumeChar(',')
		} else {
			break
		}
	}
	return selectors, nil
}

func (p *Parser) parseDeclarations() []Declaration {
	var declarations []Declaration
	for {
		p.consumeWhitespace()
		if p.peek() == '}' {
			break
		}
		property := p.consumeIdentifier()
		p.consumeWhitespace()
		if !p.consumeChar(':') {
			break // or return error
		}
		p.consumeWhitespace()
		value := p.consumeUntil(';')
		p.consumeWhitespace()
		if p.peek() == ';' {
			p.consumeChar(';')
		}

		declarations = append(declarations, Declaration{Property: property, Value: value})
		p.consumeWhitespace()
		if p.peek() == '}' {
			break
		}
	}
	return declarations
}

func (p *Parser) consumeIdentifier() string {
	var result string
	for p.pos < len(p.input) && isIdentifierChar(p.peek()) {
		result += string(p.input[p.pos])
		p.pos++
	}
	return result
}

func (p *Parser) consumeUntil(stopChar byte) string {
	var result string
	for p.pos < len(p.input) && p.peek() != stopChar {
		result += string(p.input[p.pos])
		p.pos++
	}
	return strings.TrimSpace(result)
}

func (p *Parser) consumeWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.peek()) {
		p.pos++
	}
}

func (p *Parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) consumeChar(char byte) bool {
	if p.peek() == char {
		p.pos++
		return true
	}
	return false
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t' || char == '\r'
}

func isIdentifierChar(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_'
}
