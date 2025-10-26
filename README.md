# Goosie

A minimal web browser implemented in Go using Goja (JavaScript engine), Fyne (GUI framework), and x/net/html (HTML parser).

## Features

- **HTTP Fetching**: Async fetch with cancellation support using context
- **HTML Parsing**: Parse HTML and extract body text using golang.org/x/net/html
- **HTML Rendering**: Canvas-based renderer with layout engine
  - Render tree for optimized DOM representation
  - Layout engine with box model calculations
  - Support for core HTML elements (headings, paragraphs, lists, links, images)
  - Text styling (bold, italic)
  - HTML hierarchy preservation
  - **High-performance viewport-based rendering** (30-65x faster than traditional approaches)
  - Display list caching for smooth scrolling
  - Viewport culling to only render visible content
- **Async Architecture**: Non-blocking page loads with responsive UI
  - Background goroutines for network and parsing operations
  - Loading spinner with visual feedback
  - Cancellable requests (navigate away anytime)
  - Context-based timeout and cancellation support
- **JavaScript Runtime**: Execute JavaScript with Goja engine
  - `console.log()` support
  - `document.getElementById()` support
- **GUI**: Display rendered content in a Fyne window titled "Goosie"
- **Navigation**: Full-featured navigation system
  - URL bar for entering web addresses
  - Back/Forward navigation buttons with proper state management
  - Refresh/Reload button
  - Session-based navigation history
  - Bookmark management (add/remove with visual indicators)

## Architecture

The project follows a clean architecture with the following structure:

```
goosie/
├── cmd/
│   ├── browser/          # Main GUI browser application
│   │   └── main.go
│   ├── renderer-demo/    # Renderer demo (no GUI)
│   │   └── main.go
│   └── test/             # Test/demo program (no GUI required)
│       └── main.go
├── internal/
│   ├── net/              # HTTP fetching
│   │   └── fetcher.go
│   ├── dom/              # HTML parsing
│   │   └── parser.go
│   ├── renderer/         # HTML canvas renderer
│   │   ├── node.go       # Render tree nodes
│   │   ├── layout.go     # Layout engine
│   │   ├── canvas.go     # Canvas rendering
│   │   ├── renderer.go   # Main renderer
│   │   └── README.md     # Renderer documentation
│   ├── js/               # JavaScript runtime (Goja)
│   │   └── runtime.go
│   └── ui/               # GUI rendering (Fyne)
│       └── browser.go
├── go.mod
└── README.md
```

## Dependencies

- [goja](https://github.com/dop251/goja) - JavaScript engine
- [fyne](https://fyne.io/) - Cross-platform GUI framework
- [x/net/html](https://pkg.go.dev/golang.org/x/net/html) - HTML parser

## Installation

### Prerequisites

For GUI functionality (cmd/browser), you need:

**Linux:**
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
```

**macOS:**
```bash
# Xcode command line tools
xcode-select --install
```

**Windows:**
```
# No additional dependencies required
```

### Build

```bash
# Clone the repository
git clone https://github.com/vyquocvu/goosie.git
cd goosie

# Install dependencies
go mod download

# Build the browser
go build ./cmd/browser

# Or run directly
go run ./cmd/browser
```

## Usage

### GUI Browser

Run the full browser with GUI:

```bash
go run ./cmd/browser
```

This will:
1. Open a window titled "Goosie" with navigation controls
2. Display a welcome message
3. Allow you to enter a URL in the address bar
4. Fetch and display web pages with async loading (UI stays responsive)
5. Show a loading spinner during page fetch and render
6. Enable back/forward navigation between pages
7. Support bookmark management with visual indicators
8. Initialize the Goja runtime with `console.log` and `document.getElementById`
9. Allow cancelling slow page loads by navigating to a new URL

### Testing Components (No GUI)

Test the core components without GUI dependencies:

```bash
go run ./cmd/test
```

This validates:
- HTTP fetcher
- HTML parser
- JavaScript runtime with console.log
- document.getElementById functionality

## Example

The browser demonstrates web functionality by:

1. **Navigation**: Enter URLs in the address bar to browse websites
2. **Fetching**: Downloads web pages using HTTP GET
3. **Parsing**: Parses HTML structure using golang.org/x/net/html
4. **Rendering**: Canvas-based renderer that:
   - Builds a render tree from HTML nodes
   - Calculates layout with box model
   - Renders to Fyne canvas with proper formatting
   - Supports headings, paragraphs, lists, links, and images
5. **History**: Navigate back and forward through visited pages
6. **Bookmarks**: Save and manage favorite pages with visual indicators
7. **JavaScript**: Runs JavaScript with Goja, supporting:
   ```javascript
   console.log("Page loaded: " + document.title);
   var elem = document.getElementById("main-content");
   console.log(elem.textContent);
   ```

## Development

### Project Structure

- **internal/net**: Async HTTP client with context support for fetching web pages
- **internal/dom**: HTML parser for extracting content
- **internal/renderer**: Canvas-based HTML renderer with layout engine
- **internal/js**: JavaScript runtime wrapper around Goja
- **internal/ui**: Fyne-based GUI components with loading indicator
- **cmd/browser**: Main browser application with async page loading
- **cmd/renderer-demo**: Renderer demonstration without GUI
- **cmd/test**: Testing utility without GUI dependencies

### Key Documentation

- **[ASYNC_ARCHITECTURE.md](ASYNC_ARCHITECTURE.md)**: Async fetch/render architecture
- **[PERFORMANCE.md](PERFORMANCE.md)**: Viewport culling and display list caching
- **[RENDER_ARCHITECTURE.md](RENDER_ARCHITECTURE.md)**: Multi-tree rendering system
- **[SCROLL_PERFORMANCE_SUMMARY.md](SCROLL_PERFORMANCE_SUMMARY.md)**: Scroll optimizations

### Adding Features

To add new JavaScript APIs, edit `internal/js/runtime.go`:

```go
// Add new DOM API
document.Set("querySelector", func(call goja.FunctionCall) goja.Value {
    // Implementation
})
```

To add new UI features, edit `internal/ui/browser.go`:

```go
// Add URL bar, navigation buttons, etc.
```

## Performance

Goosie includes advanced performance optimizations for smooth scrolling and high frame rates:

- **Viewport-based rendering**: 30x faster than traditional full-page rendering
- **Display list caching**: Eliminates repeated DOM traversal
- **Scroll optimization**: 65x faster scroll updates
- **Scales to thousands of elements**: Constant-time rendering regardless of page size

See [PERFORMANCE.md](PERFORMANCE.md) for detailed benchmarks and technical information.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future development goals.

## License

This project is provided as-is for educational purposes.
