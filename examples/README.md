# Render Tree Architecture Examples

This directory contains examples demonstrating the render tree and layout tree architecture.

## Architecture Overview

The rendering pipeline consists of multiple stages:

```
HTML Document → DOM Tree → Render Tree → Layout Tree → Display List → Paint
```

## Scroll Performance Demo

To see the viewport-based rendering optimizations in action:

```bash
go run examples/scroll_perf_demo.go
```

This demonstrates:
- Initial page render time
- Scroll update performance
- FPS estimation during scrolling
- Viewport culling effectiveness

Example output:
```
=== Scroll Performance Test ===

Initial render time: 357µs
Content height: 3289.92 pixels

Simulating scroll performance...
Scroll to Y=0: 5.15µs
Scroll to Y=300: 3.94µs
Scroll to Y=600: 4.33µs

Average scroll update time: 4.53µs
Estimated FPS during scrolling: 220,848

✓ Scroll performance is EXCELLENT (< 2ms per update)
```

The test page (`examples/long_page.html`) contains ~100 elements across 15 sections to simulate a real long-form web page.

## Example Usage (from internal tests)

Since the renderer package is internal, you can see examples of usage in the test files:

### Basic Usage

```go
// 1. Parse HTML to DOM
doc, err := html.Parse(strings.NewReader(htmlContent))

// 2. Build Render Tree (with unique node IDs)
renderTree := renderer.BuildRenderTree(doc)

// 3. Compute Layout Tree
layoutEngine := renderer.NewLayoutEngine(800, 600)
layoutTree := layoutEngine.ComputeLayout(renderTree)

// 4. Build Display List
displayListBuilder := renderer.NewDisplayListBuilder()
displayList := displayListBuilder.Build(layoutTree, renderTree)

// 5. Hit Testing
nodeID := layoutEngine.HitTest(layoutTree, x, y)
```

### Incremental Updates

```go
// Create incremental layout engine
ile := renderer.NewIncrementalLayoutEngine(800, 600)

// Initial layout
layoutTree := ile.ComputeIncrementalLayout(renderTree, nil)

// Mark node as dirty when it changes
ile.InvalidateNode(changedNode, renderer.DirtyLayout)

// Recompute only dirty subtrees
newLayout := ile.ComputeIncrementalLayout(renderTree, layoutTree)
```

## Running Tests

See the comprehensive tests and benchmarks:

```bash
# Run all renderer tests
go test ./internal/renderer -v

# Run benchmarks
go test ./internal/renderer -bench=. -benchmem

# View test coverage
go test ./internal/renderer -cover
```

## Example Output from Tests

The test suite demonstrates:

- **Render tree building**: Creates nodes with unique IDs and attributes
- **Layout computation**: Positions boxes with proper dimensions
- **Hit testing**: Finds deepest node at coordinates
- **Display list**: Generates paint commands for rendering
- **Invalidation**: Tracks dirty nodes for incremental updates

See `internal/renderer/*_test.go` for complete examples.

## Performance Benchmarks

From `benchmark_test.go`:

| Operation | Small (10 nodes) | Large (1000 nodes) |
|-----------|------------------|--------------------|
| Layout | ~14 μs | ~21 μs |
| HitTest | ~26 ns | ~26 ns |
| DisplayList | ~1 μs | ~2 μs |
| Full Pipeline | ~15 μs | ~23 μs |
| **Viewport Rendering** | **~413 ns** | **~746 ns** |
| **Viewport Scroll** | **~179 ns** | **~350 ns** |

**Key Insights:**
- Viewport rendering is **30x faster** than full pipeline
- Scroll updates are **65x faster** than full pipeline
- Performance is constant regardless of page size (viewport culling)

## Documentation

For detailed architecture documentation, see:
- [RENDER_ARCHITECTURE.md](../RENDER_ARCHITECTURE.md) - Complete architecture guide
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Overall system architecture
