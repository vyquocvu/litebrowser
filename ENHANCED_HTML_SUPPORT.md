# Enhanced HTML Support

This document describes the enhanced HTML support features in Goosie v0.5.0.

## Overview

Goosie now supports comprehensive HTML rendering with CSS styling, form elements, tables, and multiple image formats. These features make Goosie capable of rendering modern web pages with proper styling and interactivity.

## Features

### 1. CSS Styling Support

Goosie supports basic CSS styling through `<style>` tags in HTML documents. The following CSS properties are supported:

#### Supported CSS Properties

- **color**: Text color (named colors and hex codes)
  - Named colors: `red`, `green`, `blue`, `black`, `white`, `yellow`, `cyan`, `magenta`, `silver`, `gray`, `maroon`, `olive`, `purple`, `teal`, `navy`
  - Hex codes: `#ff0000`, `#00ff00`, `#0000ff`, etc.
  - Short hex: `#f00`, `#0f0`, `#00f`, etc.

- **font-size**: Text size
  - Pixel units: `12px`, `16px`, `24px`, etc.
  - Em units: `0.8em`, `1.2em`, `1.5em`, etc. (relative to parent)

- **font-weight**: Text weight
  - Values: `normal`, `bold`

- **font-family**: Font family name (parsed but not fully applied due to Fyne limitations)

- **background-color**: Background color (parsed but not rendered in current implementation)

#### CSS Selectors

Goosie supports the following CSS selectors:

- **Element selector**: `p { color: red; }`
- **Class selector**: `.my-class { color: blue; }`
- **ID selector**: `#my-id { font-size: 20px; }`
- **Pseudo-class selectors**: `:link`, `:visited` (for links)

#### Example

```html
<!DOCTYPE html>
<html>
<head>
    <style>
        .red-text {
            color: red;
            font-size: 18px;
        }
        
        .blue-header {
            color: #0000ff;
            font-size: 24px;
            font-weight: bold;
        }
        
        p {
            color: #333333;
        }
    </style>
</head>
<body>
    <h1 class="blue-header">Welcome to Goosie</h1>
    <p class="red-text">This text is red and 18px.</p>
    <p>This text is dark gray (default paragraph color).</p>
</body>
</html>
```

### 2. Image Rendering

Goosie supports loading and rendering images in multiple formats:

#### Supported Formats

- **PNG**: Portable Network Graphics (with transparency support)
- **JPEG**: Joint Photographic Experts Group
- **GIF**: Graphics Interchange Format (including animated GIFs - first frame only)
- **WebP**: Modern image format with compression

#### Features

- Async image loading (non-blocking)
- Image caching (LRU cache with configurable size)
- Relative and absolute URL resolution
- Alt text display when image fails to load
- Loading indicators

#### Example

```html
<img src="https://example.com/image.png" alt="Example image" />
<img src="photo.jpg" alt="Photo" />
<img src="animation.gif" alt="Animation" />
<img src="modern.webp" alt="WebP image" />
```

### 3. Interactive Links

All links (`<a>` tags) are interactive and clickable. When clicked, they trigger navigation to the target URL.

#### Features

- Absolute URL navigation
- Relative URL resolution
- Navigation callback system
- Proper URL parsing and validation

#### Example

```html
<a href="https://example.com">Visit Example.com</a>
<a href="/about">About page (relative)</a>
<a href="../parent/page.html">Parent directory</a>
```

### 4. Form Elements

Goosie renders standard HTML form elements with basic interactivity:

#### Supported Elements

- **Input**: `<input>` - Single-line text input with placeholder support
- **Button**: `<button>` - Clickable button with label
- **Textarea**: `<textarea>` - Multi-line text input with placeholder support

#### Example

```html
<form>
    <input type="text" placeholder="Enter your name" />
    <input type="email" placeholder="Email address" />
    <textarea placeholder="Enter your message"></textarea>
    <button>Submit</button>
</form>
```

### 5. Table Rendering

Tables are rendered using Fyne's Table widget with proper layout and cell management.

#### Features

- Support for `<table>`, `<tbody>`, `<thead>`, `<tfoot>`, `<tr>`, `<td>`, `<th>`
- Automatic column width management
- Text extraction from cells
- Nested content support

#### Example

```html
<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Age</th>
            <th>City</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>Alice</td>
            <td>30</td>
            <td>New York</td>
        </tr>
        <tr>
            <td>Bob</td>
            <td>25</td>
            <td>Los Angeles</td>
        </tr>
    </tbody>
</table>
```

## Technical Implementation

### CSS Parser

The CSS parser (`internal/css/parser.go`) processes CSS text and builds a stylesheet with rules and declarations. The parser supports:

- Selector parsing (element, class, ID, pseudo-class)
- Declaration parsing (property-value pairs)
- Multiple selectors per rule
- Comments (ignored)

### Style Manager

The Style Manager (`internal/renderer/style.go`) applies CSS rules to the render tree:

1. Matches selectors against nodes
2. Applies declarations to matched nodes
3. Handles style inheritance
4. Stores computed styles in each node

### Canvas Renderer

The Canvas Renderer (`internal/renderer/canvas.go`) renders styled content:

- Uses `canvas.Text` objects for CSS-styled text (supports custom colors and sizes)
- Uses `widget.Label` for non-styled text (better wrapping support)
- Checks `ComputedStyle` for each node during rendering
- Falls back to default styles when no CSS is present

### Image Loader

The Image Loader (`internal/image/loader.go`) handles image loading:

- Async loading with goroutines
- LRU caching to reduce network requests
- Support for multiple formats via Go's `image` package
- Callback system for refresh after load

## Limitations

### Current Limitations

1. **Font sizes**: Custom font sizes work but may not wrap properly due to Fyne limitations
2. **Background colors**: Parsed but not rendered (Fyne widget limitation)
3. **Advanced CSS**: No support for:
   - Flexbox or Grid layouts
   - Margins and padding (partially supported in layout engine)
   - Borders
   - Animations
   - Transitions
   - Media queries
4. **Form submission**: Forms are rendered but don't submit data
5. **Table styling**: Tables use fixed column widths, no CSS styling

### Known Issues

- Text wrapping may be imperfect for CSS-styled text using `canvas.Text`
- Some Fyne themes may override default colors
- Very large images may affect performance

## Examples

See the `examples/enhanced_html_demo.html` file for a comprehensive demonstration of all enhanced HTML features.

## Testing

Comprehensive tests are available in:

- `internal/renderer/enhanced_html_test.go` - CSS styling tests
- `internal/renderer/form_and_table_test.go` - Form and table tests
- `internal/image/loader_test.go` - Image loading tests
- `internal/css/parser_test.go` - CSS parser tests

Run tests with:

```bash
go test ./internal/renderer -v
go test ./internal/css -v
go test ./internal/image -v
```

## Future Enhancements

Planned improvements for future versions:

- Full CSS box model support (margins, padding, borders)
- CSS Flexbox and Grid layouts
- Advanced CSS selectors (descendant, child, sibling)
- More CSS properties (text-align, text-decoration, etc.)
- Form submission and validation
- Table CSS styling
- SVG rendering
- Canvas API support

## API Reference

### Setting CSS Styles Programmatically

While CSS is typically parsed from `<style>` tags, you can also programmatically set styles:

```go
node := renderer.NewRenderNode(renderer.NodeTypeElement)
node.ComputedStyle = &renderer.Style{
    Color:      color.RGBA{R: 255, G: 0, B: 0, A: 255},
    FontSize:   18.0,
    FontWeight: "bold",
}
```

### Custom Navigation Callback

Set a custom callback for link clicks:

```go
renderer := renderer.NewRenderer(800, 600)
renderer.SetNavigationCallback(func(url string) {
    fmt.Printf("Navigating to: %s\n", url)
    // Handle navigation
})
```

### Image Cache Configuration

Configure the image cache size:

```go
imageLoader := imageloader.NewLoader(100) // Cache up to 100 images
```

## Changelog

### v0.5.0 (November 2025)

- ✅ Added CSS basic styling support (colors, fonts, sizes)
- ✅ Added full image rendering (PNG, JPEG, GIF, WebP)
- ✅ Added interactive link click handling
- ✅ Added form elements rendering (input, button, textarea)
- ✅ Added table rendering with tbody/thead/tfoot support
- ✅ Improved CSS parser with support for hex colors
- ✅ Enhanced display list rendering with CSS style application

---

*For more information, see the main [README.md](README.md) and [ROADMAP.md](ROADMAP.md).*
