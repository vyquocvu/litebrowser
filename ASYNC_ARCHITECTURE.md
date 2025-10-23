# Async Fetch/Render Architecture

This document describes the async fetch and render architecture implemented in Goosie to prevent UI freezing during page loads.

## Problem Statement

Previously, Goosie's UI would freeze during page fetch and render operations because all work happened on the main UI thread. This created poor user experience, especially on slow networks or with large HTML documents.

## Solution Overview

We've implemented an asynchronous architecture that:
1. **Moves network and parsing to background goroutines** - UI remains responsive
2. **Shows loading indicator** - Users see visual feedback during operations
3. **Supports cancellation** - Users can navigate away from slow-loading pages
4. **Preserves all performance optimizations** - Viewport culling, display list caching, etc.

## Architecture

### Before (Blocking)

```
User navigates to URL
   ↓
[Main UI Thread - BLOCKED]
 - HTTP fetch (blocks UI)
 - HTML parsing (blocks UI)
 - Render tree/layout (blocks UI)
 - Canvas rendering (blocks UI)
   ↓
UI responds only after all work completes
```

**Problems:**
- UI frozen during entire operation
- No progress feedback
- Cannot cancel slow loads
- Poor UX on slow networks

### After (Async)

```
User navigates to URL
   ↓
[Main UI Thread - RESPONSIVE]
 - Show loading spinner
 - Update URL bar
 - Update navigation buttons
   ↓
[Background Goroutine]
 - HTTP fetch
 - HTML parsing  
 - Check for cancellation
   ↓
[Main UI Thread - Update]
 - Render tree/layout
 - Canvas rendering
 - Hide loading spinner
```

**Benefits:**
- UI never freezes
- Loading indicator provides feedback
- Users can cancel or navigate during load
- Responsive even on slow connections

## Implementation Details

### 1. Context-based HTTP Fetching

**File:** `internal/net/fetcher.go`

Added `FetchWithContext` method that accepts a `context.Context`:

```go
func (f *Fetcher) FetchWithContext(ctx context.Context, url string) (string, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    // ... handle request with context
}
```

**Key features:**
- Cancellable HTTP requests
- Timeout support via context
- Backward compatible (old `Fetch()` calls `FetchWithContext` with background context)

### 2. Loading Indicator UI

**File:** `internal/ui/browser.go`

Added UI components for loading feedback:

```go
type Browser struct {
    // ... existing fields
    loadingSpinner *widget.ProgressBarInfinite
    loadingLabel   *widget.Label
}

func (b *Browser) ShowLoading() {
    b.loadingSpinner.Show()
    b.loadingLabel.Show()
    b.loadingSpinner.Start()
}

func (b *Browser) HideLoading() {
    b.loadingSpinner.Stop()
    b.loadingSpinner.Hide()
    b.loadingLabel.Hide()
}
```

**UI Layout:**
```
┌─────────────────────────────────────┐
│ [←] [→] [⟳] [URL Entry]      [☆]   │ ← Navigation Bar
├─────────────────────────────────────┤
│ [●●●] Loading...                    │ ← Loading Indicator (shown during load)
├─────────────────────────────────────┤
│                                     │
│         Page Content                │ ← Scrollable Content Area
│                                     │
└─────────────────────────────────────┘
```

### 3. Async Page Loading

**File:** `cmd/browser/main.go`

Implemented async page loading with cancellation:

```go
func main() {
    var currentLoadCtx context.Context
    var currentLoadCancel context.CancelFunc

    browser.SetNavigationCallback(func(url string) {
        // Cancel any ongoing page load
        if currentLoadCancel != nil {
            currentLoadCancel()
        }

        // Create new context for this load
        currentLoadCtx, currentLoadCancel = context.WithCancel(context.Background())

        // Load page asynchronously
        loadPageAsync(browser, fetcher, parser, jsRuntime, url, currentLoadCtx)
    })
}
```

**Async load function:**

```go
func loadPageAsync(browser *ui.Browser, fetcher *net.Fetcher, 
                   parser *dom.Parser, jsRuntime *js.Runtime, 
                   url string, ctx context.Context) {
    
    // Update UI state (on main thread)
    browser.NavigateTo(url)
    browser.ShowLoading()

    // Background goroutine for network I/O
    go func() {
        // Fetch with cancellation support
        html, err := fetcher.FetchWithContext(ctx, url)
        
        // Check if cancelled
        if ctx.Err() != nil {
            log.Printf("Page load cancelled for: %s", url)
            return
        }

        // Handle errors or update UI with content
        if err != nil {
            updateUIWithError(browser, err)
        } else {
            updateUIWithContent(browser, jsRuntime, html, url)
        }
    }()
}
```

## Cancellation Flow

When a user rapidly navigates between pages:

```
Time 0ms:  User navigates to URL1
           → Context1 created
           → Fetch URL1 starts in background
           → Spinner shown

Time 100ms: User navigates to URL2
           → Context1 cancelled (aborts URL1 fetch)
           → Context2 created
           → Fetch URL2 starts in background
           
Time 150ms: URL1 fetch detects cancellation
           → Returns early, no UI update

Time 300ms: URL2 fetch completes
           → UI updated with URL2 content
           → Spinner hidden
```

## Thread Safety

Fyne framework provides thread-safe widget updates, so UI methods can be called from any goroutine:

- `browser.SetContent()` - thread-safe
- `browser.RenderHTMLContent()` - thread-safe
- `browser.ShowLoading() / HideLoading()` - thread-safe

All UI updates happen through Fyne's internal synchronization mechanisms.

## Performance Considerations

### No Impact on Rendering Performance

The async architecture has **zero impact** on rendering performance because:

1. **Viewport culling** - Still renders only visible elements (see PERFORMANCE.md)
2. **Display list caching** - Still caches paint commands for scroll
3. **Incremental layout** - Still uses invalidation tracking
4. **Canvas optimization** - All optimizations preserved

The only difference is **when** rendering happens (in background goroutine vs main thread).

### Benefits

- **Responsive UI**: Main thread never blocked
- **Better perceived performance**: Loading indicator shows progress
- **Improved UX**: Users can cancel slow loads
- **Network efficiency**: Cancelled requests stop immediately

## Testing

### Unit Tests

**File:** `internal/net/fetcher_test.go`

Basic context support:
- `TestFetchWithContextCancellation` - Immediate cancellation
- `TestFetchWithContextTimeout` - Timeout handling

### Integration Tests

**File:** `internal/net/async_test.go`

Real-world async scenarios:
- `TestAsyncFetchCancellation` - Cancel mid-flight request
- `TestConcurrentFetches` - Multiple simultaneous fetches
- `TestAsyncFetchWithTimeout` - Timeout with slow server
- `TestMultipleNavigationCancellation` - Rapid navigation between pages

All tests validate:
- Proper cancellation handling
- No goroutine leaks
- Correct error propagation
- Thread safety

## Migration Guide

### For Existing Code

No changes needed! The implementation is backward compatible:

```go
// Old code still works
html, err := fetcher.Fetch(url)

// New code can use context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
html, err := fetcher.FetchWithContext(ctx, url)
```

### For New Features

When adding new network operations:

1. **Use context**: Accept `context.Context` parameter
2. **Check cancellation**: Use `ctx.Err()` to detect cancellation
3. **Show progress**: Use `ShowLoading()` / `HideLoading()`
4. **Update async**: Use goroutines for I/O, thread-safe UI updates

## Future Enhancements

Possible improvements to async architecture:

1. **Streaming parsing** - Start rendering before full page loads
2. **Worker pool** - Limit concurrent background operations
3. **Progress bar** - Show bytes downloaded / DOM parsed
4. **Priority queue** - User actions prioritized over background loads
5. **Smart caching** - Cache parsed DOM for faster back/forward
6. **Prefetching** - Speculatively fetch linked pages

## References

- **PERFORMANCE.md** - Viewport culling, display list caching
- **RENDER_ARCHITECTURE.md** - Multi-tree render architecture
- **SCROLL_PERFORMANCE_SUMMARY.md** - Scroll optimizations
- **Go Blog: Context** - https://go.dev/blog/context
- **Fyne Threading** - https://developer.fyne.io/architecture/threads

## Acceptance Criteria

✅ UI never blocks during fetch/render
✅ Spinner/progress visible during load
✅ User can cancel or navigate during load
✅ No regression in scroll/render performance
✅ All tests continue to pass
