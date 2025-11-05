# Full CSS Parser Implementation

This document describes the full CSS parser implementation in Goosie, which provides comprehensive support for CSS selectors, combinators, pseudo-classes, pseudo-elements, and at-rules.

## Overview

The CSS parser is located in `internal/css/` and consists of:
- `parser.go` - The main CSS parser
- `stylesheet.go` - CSS data structures
- `parser_test.go` - Comprehensive test suite

The style matching implementation is in `internal/renderer/style.go`.

## Features

### 1. Basic Selectors

- **Tag selectors**: `h1`, `p`, `div`
- **Class selectors**: `.classname`, `.multiple.classes`
- **ID selectors**: `#elementid`
- **Universal selector**: `*`
- **Combined selectors**: `div.classname#id`

### 2. Combinators

- **Descendant combinator** (space): `div p` - Matches any `<p>` inside a `<div>` at any level
- **Child combinator** (`>`): `ul > li` - Matches `<li>` that is a direct child of `<ul>`
- **Adjacent sibling combinator** (`+`): `h1 + p` - Matches `<p>` immediately following `<h1>`
- **General sibling combinator** (`~`): `h1 ~ p` - Matches any `<p>` that follows `<h1>` as a sibling

### 3. Attribute Selectors

- **Presence**: `[disabled]` - Matches elements with the attribute
- **Exact match** (`=`): `[type="text"]` - Matches exact value
- **Word match** (`~=`): `[class~="active"]` - Matches if value contains the word
- **Prefix match** (`|=`): `[lang|="en"]` - Matches exact value or value followed by hyphen
- **Starts with** (`^=`): `[href^="https"]` - Matches if value starts with substring
- **Ends with** (`$=`): `[src$=".png"]` - Matches if value ends with substring
- **Contains** (`*=`): `[title*="example"]` - Matches if value contains substring

### 4. Pseudo-Classes

- `:link` - Unvisited links
- `:visited` - Visited links
- `:hover` - Element being hovered (state tracking not fully implemented)
- `:focus` - Element with focus (state tracking not fully implemented)
- `:active` - Active element (state tracking not fully implemented)
- `:first-child` - First child of parent
- `:last-child` - Last child of parent
- `:nth-child(n)` - Nth child of parent (basic support)

### 5. Pseudo-Elements

- `::before` - Insert content before element
- `::after` - Insert content after element
- Other pseudo-elements are parsed but not yet rendered

### 6. Multiple Selectors

Comma-separated selectors:
```css
h1, h2, h3 {
    font-family: Arial;
}
```

### 7. CSS Comments

Standard CSS comments are supported:
```css
/* This is a comment */
selector { /* inline comment */
    property: value;
}
```

### 8. At-Rules

- `@media` - Media queries (parsed, conditionals not evaluated yet)
- `@import` - Import external stylesheets (parsed, not fetched)
- `@keyframes` - Animation keyframes (parsed, animations not implemented)
- `@supports` - Feature queries (parsed, not evaluated)

### 9. Important Flag

The `!important` flag is supported:
```css
.override {
    color: red !important;
}
```

### 10. Complex Values

- **Functions**: `url()`, `calc()`, `rgb()`, `rgba()` - Parsed and preserved
- **Quoted strings**: Both single and double quotes
- **Multiple values**: Space-separated values
- **Escape sequences**: Backslash escapes in strings

## Architecture

### Data Structures

```go
// StyleSheet - Root structure
type StyleSheet struct {
    Rules   []Rule
    AtRules []AtRule
}

// Rule - A CSS rule with selectors and declarations
type Rule struct {
    Selectors    []SelectorSequence
    Declarations []Declaration
}

// SelectorSequence - A selector with optional combinators
// e.g., "div > p.class" is represented as:
//   Simple: {TagName: "div"}
//   Combinator: ">"
//   Next: {Simple: {TagName: "p", Classes: ["class"]}}
type SelectorSequence struct {
    Simple     SimpleSelector
    Combinator string // "", " ", ">", "+", "~"
    Next       *SelectorSequence
}

// SimpleSelector - A simple selector (no combinators)
type SimpleSelector struct {
    TagName        string
    ID             string
    Classes        []string
    PseudoClasses  []string
    PseudoElements []string
    Attributes     []AttributeSelector
    Universal      bool
}

// Declaration - A property-value pair
type Declaration struct {
    Property  string
    Value     string
    Important bool
}
```

### Parser Algorithm

1. **Tokenization**: The parser reads the CSS input character by character
2. **Selector parsing**: Parses selectors left-to-right, building a linked list
3. **Declaration parsing**: Parses property-value pairs, handling nested functions
4. **Comment stripping**: Removes CSS comments during parsing
5. **At-rule handling**: Identifies and parses at-rules separately

### Matching Algorithm

The style matcher uses a **right-to-left** matching approach:

1. For selector sequence `A > B`:
   - First check if the current node matches `B`
   - Then check if the parent matches `A`
   
2. For descendant selector `A B`:
   - Check if current node matches `B`
   - Walk up the tree to find an ancestor matching `A`

3. For sibling selectors:
   - Check if current node matches the rightmost selector
   - Search siblings for matching elements

This approach is efficient because most selector mismatches are caught early.

## Examples

### Complex Selector Example

```css
div.container > p#intro.highlight:first-child {
    color: blue;
    font-size: 18px;
}
```

This matches a `<p>` element that:
- Has ID "intro"
- Has class "highlight"
- Is the first child of its parent
- Is a direct child of a `<div>` with class "container"

### Attribute Selector Example

```css
/* Match all HTTPS links */
a[href^="https"] {
    color: green;
}

/* Match all PNG images */
img[src$=".png"] {
    border: 1px solid gray;
}

/* Match elements with 'active' in their class */
[class~="active"] {
    font-weight: bold;
}
```

### Sibling Selector Example

```css
/* Style the first paragraph after each h2 */
h2 + p {
    font-size: 18px;
    margin-top: 0;
}

/* Style all paragraphs after an h2 */
h2 ~ p {
    color: #555;
}
```

## Testing

The parser includes comprehensive tests in `parser_test.go`:

- `TestParser` - Basic selector and declaration parsing
- `TestParserCombinedSelector` - Class and tag combinations
- `TestParserComments` - Comment handling
- `TestParserDescendantSelector` - Descendant combinator
- `TestParserChildSelector` - Child combinator
- `TestParserAdjacentSiblingSelector` - Adjacent sibling
- `TestParserGeneralSiblingSelector` - General sibling
- `TestParserAttributeSelector` - All attribute selector types
- `TestParserUniversalSelector` - Universal selector
- `TestParserPseudoClasses` - Pseudo-class support
- `TestParserPseudoElements` - Pseudo-element support
- `TestParserMultipleSelectors` - Comma-separated selectors
- `TestParserImportant` - !important flag
- `TestParserAtMedia` - @media rules
- `TestParserComplexSelector` - Complex multi-part selectors
- `TestParserValueWithFunction` - Function values

Run tests with:
```bash
go test ./internal/css/... -v
```

## Limitations and Future Work

### Current Limitations

1. **Pseudo-class state**: `:hover`, `:focus`, `:active` require UI state tracking
2. **Nth-child logic**: Only basic support, complex formulas not yet implemented
3. **Media queries**: Parsed but conditions not evaluated
4. **Pseudo-elements**: Parsed but content generation not implemented
5. **Specificity**: Not yet implemented for cascade resolution
6. **Cascade and inheritance**: Basic inheritance, full cascade not implemented

### Future Enhancements

1. Implement specificity calculation for proper cascade
2. Add support for CSS variables (custom properties)
3. Implement pseudo-element content generation
4. Add media query evaluation
5. Support more pseudo-classes (`:not()`, `:is()`, `:where()`)
6. Implement shorthand property expansion (margin, padding, border)
7. Add support for CSS Grid and Flexbox layout

## Demo

See `examples/html/full_css_demo.html` for a comprehensive demonstration of all supported CSS features.

## Related Files

- `internal/css/parser.go` - Main parser implementation
- `internal/css/stylesheet.go` - Data structures
- `internal/css/parser_test.go` - Test suite
- `internal/renderer/style.go` - Style matching and application
- `examples/html/full_css_demo.html` - Feature demonstration
