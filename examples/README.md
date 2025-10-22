# Render Tree Architecture Examples

This directory contains examples demonstrating the render tree and layout tree architecture.

## Architecture Overview

The rendering pipeline consists of multiple stages:

```
HTML Document → DOM Tree → Render Tree → Layout Tree → Display List → Paint
```

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
| Layout | ~14 μs | ~22 μs |
| HitTest | ~26 ns | ~26 ns |
| DisplayList | ~1 μs | ~2 μs |
| Full Pipeline | ~15 μs | ~23 μs |

## Documentation

For detailed architecture documentation, see:
- [RENDER_ARCHITECTURE.md](../RENDER_ARCHITECTURE.md) - Complete architecture guide
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Overall system architecture
