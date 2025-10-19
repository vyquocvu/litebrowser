# litebrowser

A minimal web browser implemented in Go using Goja (JavaScript engine), Fyne (GUI framework), and x/net/html (HTML parser).

## Features

- **HTTP Fetching**: Fetch web pages using the built-in net/http package
- **HTML Parsing**: Parse HTML and extract body text using golang.org/x/net/html
- **JavaScript Runtime**: Execute JavaScript with Goja engine
  - `console.log()` support
  - `document.getElementById()` support
- **GUI**: Display rendered content in a Fyne window titled "Goja Browser"

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
1. Open a window titled "Goja Browser"
2. Fetch https://example.com
3. Parse the `<body>` text content
4. Render it in the canvas
5. Initialize the Goja runtime with `console.log` and `document.getElementById`
6. Run a test JS snippet logging "JS runtime initialized"

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

The browser demonstrates basic web functionality by:

1. **Fetching**: Downloads https://example.com using HTTP GET
2. **Parsing**: Extracts text from the `<body>` element
3. **Rendering**: Displays the text in a scrollable Fyne canvas
4. **JavaScript**: Runs JavaScript with Goja, supporting:
   ```javascript
   console.log("JS runtime initialized");
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

## License

This project is provided as-is for educational purposes.
