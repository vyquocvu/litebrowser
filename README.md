# litebrowser

A minimal web browser implemented in Go using Goja (JavaScript engine), Fyne (GUI framework), and x/net/html (HTML parser).

## Features

- **HTTP Fetching**: Fetch web pages using the built-in net/http package
- **HTML Parsing**: Parse HTML and extract body text using golang.org/x/net/html
- **JavaScript Runtime**: Execute JavaScript with Goja engine
  - `console.log()` support
  - `document.getElementById()` support
- **GUI**: Display rendered content in a Fyne window titled "Litebrowser"
- **Navigation**: Full-featured navigation system
  - URL bar for entering web addresses
  - Back/Forward navigation buttons with proper state management
  - Refresh/Reload button
  - Session-based navigation history
  - Bookmark management (add/remove with visual indicators)

## Architecture

The project follows a clean architecture with the following structure:

```
litebrowser/
├── cmd/
│   ├── browser/          # Main GUI browser application
│   │   └── main.go
│   └── test/             # Test/demo program (no GUI required)
│       └── main.go
├── internal/
│   ├── net/              # HTTP fetching
│   │   └── fetcher.go
│   ├── dom/              # HTML parsing
│   │   └── parser.go
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
git clone https://github.com/vyquocvu/litebrowser.git
cd litebrowser

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
1. Open a window titled "Litebrowser" with navigation controls
2. Display a welcome message
3. Allow you to enter a URL in the address bar
4. Fetch and display web pages with full navigation support
5. Enable back/forward navigation between pages
6. Support bookmark management with visual indicators
7. Initialize the Goja runtime with `console.log` and `document.getElementById`

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
3. **Parsing**: Extracts text from the `<body>` element and converts to markdown
4. **Rendering**: Displays the content in a scrollable Fyne canvas
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

- **internal/net**: HTTP client for fetching web pages
- **internal/dom**: HTML parser for extracting content
- **internal/js**: JavaScript runtime wrapper around Goja
- **internal/ui**: Fyne-based GUI components
- **cmd/browser**: Main browser application
- **cmd/test**: Testing utility without GUI dependencies

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

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future development goals.

## License

This project is provided as-is for educational purposes.
