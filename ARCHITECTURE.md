# Litebrowser Architecture Diagram

## Component Flow

```
┌─────────────────────────────────────────────────────────────┐
│                        Main Browser                          │
│                    (cmd/browser/main.go)                     │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   HTTP       │   │   HTML       │   │ JavaScript   │
│   Fetcher    │──▶│   Parser     │──▶│   Runtime    │
│ (internal/   │   │ (internal/   │   │ (internal/   │
│    net)      │   │    dom)      │   │    js)       │
└──────────────┘   └──────────────┘   └──────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                            ▼
                   ┌──────────────┐
                   │  GUI Browser │
                   │ (internal/   │
                   │    ui)       │
                   │ [Fyne Window]│
                   └──────────────┘
```

## Example Execution Flow

1. **HTTP Fetcher** (`internal/net/fetcher.go`)
   - Fetches https://example.com
   - Returns HTML content

2. **HTML Parser** (`internal/dom/parser.go`)
   - Parses HTML using x/net/html
   - Extracts body text: "Example Domain This domain is for use..."
   - Provides getElementById functionality

3. **JavaScript Runtime** (`internal/js/runtime.go`)
   - Initializes Goja VM
   - Sets up console.log API
   - Sets up document.getElementById API
   - Runs: `console.log("JS runtime initialized")`
   - Output: "JS runtime initialized"

4. **GUI Browser** (`internal/ui/browser.go`)
   - Creates Fyne window titled "Goja Browser"
   - Displays parsed body text in scrollable canvas
   - Renders the content

## Window Layout (when GUI is available)

```
┌─────────────────────────────────────────────────┐
│ Goja Browser                              [_][□][X]│
├─────────────────────────────────────────────────┤
│                                                 │
│  Example Domain                                 │
│                                                 │
│  This domain is for use in illustrative        │
│  examples in documents. You may use this        │
│  domain in literature without prior             │
│  coordination or asking for permission.         │
│                                                 │
│  More information...                            │
│                                                 │
│                                                 │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Console Output

```
2025/10/19 18:04:49 Fetching https://example.com...
2025/10/19 18:04:49 Parsed body text: Example Domain This domain...
2025/10/19 18:04:49 Running test JavaScript...
JS runtime initialized
```

## Key Features Demonstrated

✓ HTTP fetching with Go's net/http
✓ HTML parsing with golang.org/x/net/html  
✓ JavaScript execution with Goja
✓ console.log() implementation
✓ document.getElementById() implementation
✓ GUI rendering with Fyne
✓ Window titled "Goja Browser"
✓ Body text extraction and display
✓ JavaScript runtime initialization

## Testing

Run without GUI (headless testing):
```bash
go run ./cmd/test
```

Run with GUI (requires X11/Wayland):
```bash
go run ./cmd/browser
```

Unit tests:
```bash
go test -v -cover ./internal/...
```

## Test Coverage

- internal/net: 36.4%
- internal/dom: 95.0%  
- internal/js: 92.9%
