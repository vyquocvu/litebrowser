# Inline Layout Engine

This document describes the inline layout engine implementation in the litebrowser renderer.

## Overview

The inline layout engine handles the layout of inline and inline-block elements according to HTML/CSS specifications. It implements proper line breaking, white space handling, and vertical alignment.

## Architecture

### Key Components

#### 1. InlineLayoutEngine
The main engine that coordinates inline layout:
- Processes inline content (text and inline elements)
- Creates line boxes for horizontal layout
- Handles word wrapping and line breaking
- Manages white space processing

#### 2. LineBox
Represents a horizontal line containing inline elements:
- **X, Y**: Position of the line (Y is baseline position)
- **Width, Height**: Dimensions of the line
- **Ascent, Descent**: Metrics from baseline to top/bottom
- **InlineBoxes**: List of inline boxes on this line
- **AvailableWidth**: Maximum width for content

#### 3. InlineBox
Represents an inline-level box (text or inline element):
- **NodeID**: Reference to RenderNode
- **X, Y**: Position relative to line
- **Width, Height**: Dimensions
- **Ascent, Descent**: Baseline metrics
- **Text**: Text content (for text nodes)
- **IsText**: Whether this is a text node
- **VerticalAlign**: Vertical alignment mode

### White Space Processing

The engine supports all CSS white-space modes:

1. **WhiteSpaceNormal** (default)
   - Collapses sequences of white space into single spaces
   - Allows text wrapping at word boundaries
   
2. **WhiteSpaceNoWrap**
   - Collapses white space like normal
   - Prevents text wrapping
   
3. **WhiteSpacePre**
   - Preserves all white space exactly as written
   - Prevents text wrapping
   
4. **WhiteSpacePreWrap**
   - Preserves white space
   - Allows text wrapping at white space characters
   
5. **WhiteSpacePreLine**
   - Collapses white space except for newlines
   - Allows text wrapping

### Line Breaking

The engine implements sophisticated line breaking:

#### Word Wrapping
- Text is split into words at white space boundaries
- Words are added to lines until they don't fit
- When a word doesn't fit, a new line is started

#### Character Breaking
- Very long words that don't fit on a single line are broken at character boundaries
- This prevents overflow and ensures content is always visible
- Characters are added one by one until the line is full

### Vertical Alignment

The engine supports all CSS vertical-align values:

1. **VerticalAlignBaseline** (default)
   - Aligns box baseline with parent baseline
   
2. **VerticalAlignTop**
   - Aligns box top with line box top
   
3. **VerticalAlignBottom**
   - Aligns box bottom with line box bottom
   
4. **VerticalAlignMiddle**
   - Centers box vertically in line box
   
5. **VerticalAlignTextTop**
   - Aligns box top with parent content area top
   
6. **VerticalAlignTextBottom**
   - Aligns box bottom with parent content area bottom
   
7. **VerticalAlignSub**
   - Subscript alignment (below baseline)
   
8. **VerticalAlignSuper**
   - Superscript alignment (above baseline)

### Inline-Block Support

The engine recognizes inline-block elements:
- `img`, `button`, `input`, `select`
- These elements participate in inline flow but have block-like properties
- They are positioned on the baseline like text

## Usage

### Basic Layout

```go
// Create inline layout engine
fontMetrics := NewFontMetrics(16.0)
ile := NewInlineLayoutEngine(fontMetrics, 16.0)

// Layout inline content
lines, totalHeight := ile.LayoutInlineContent(
    node,               // RenderNode with inline children
    0,                  // X position
    0,                  // Y position
    400,                // Available width
    WhiteSpaceNormal,   // White space mode
)

// Process lines
for _, line := range lines {
    for _, inlineBox := range line.InlineBoxes {
        // Render inline box at line.X + inlineBox.X, line.Y + inlineBox.Y
    }
}
```

### Integration with Layout Engine

The inline layout engine is automatically used by the main LayoutEngine:

```go
// Create layout engine
le := NewLayoutEngine(800, 600)

// Compute layout (automatically uses inline layout for inline content)
layoutRoot := le.ComputeLayout(renderRoot)

// Access line boxes if needed
for _, layoutBox := range layoutRoot.Children {
    if len(layoutBox.LineBoxes) > 0 {
        // This box contains inline content with line boxes
        for _, line := range layoutBox.LineBoxes {
            // Process line
        }
    }
}
```

## Performance Optimizations

### 1. Efficient Text Measurement
- Uses cached font metrics when available
- Falls back to character-based estimation in test environments

### 2. Smart Line Breaking
- Word wrapping minimizes line changes
- Character breaking only when necessary
- Avoids unnecessary text remeasurement

### 3. Incremental Layout
- Line boxes are created incrementally
- Previous content is not recalculated
- Minimal memory allocation

### 4. Viewport Optimization
- Works with display list caching
- Only visible content needs to be rendered
- Line boxes support efficient hit testing

## Testing

The inline layout engine has comprehensive test coverage:

### Unit Tests
- White space processing (all modes)
- Text splitting and word breaking
- Character-level breaking
- Line box creation and finalization
- Vertical alignment
- Inline-block detection

### Integration Tests
- Simple paragraphs
- Long text with wrapping
- Multiple text nodes
- Mixed inline elements
- Empty/whitespace-only content

Run tests:
```bash
go test ./internal/renderer/... -run Inline -v
```

## Future Enhancements

### Planned Features
1. **Bidirectional Text Support**
   - Right-to-left text (Arabic, Hebrew)
   - Mixed directionality

2. **Advanced Typography**
   - Hyphenation
   - Justification
   - Ligatures

3. **Inline Formatting Context**
   - Full CSS inline formatting model
   - Anonymous inline boxes
   - Strut calculation

4. **Performance**
   - Text shaping cache
   - Line break cache
   - Lazy layout for off-screen content

## References

- [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/)
- [CSS Inline Layout Module](https://www.w3.org/TR/css-inline-3/)
- [CSS Box Model](https://www.w3.org/TR/css-box-3/)

## License

This module is part of the litebrowser project and follows the same license.
