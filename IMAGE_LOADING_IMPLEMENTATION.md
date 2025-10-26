# Image Loading Implementation

This document describes the implementation of image loading, decoding, and caching in the Goosie project.

## Overview

The image loading system provides:
- Loading images from remote URLs (HTTP/HTTPS)
- Loading images from local file paths
- Support for common image formats (PNG, JPEG, GIF, WebP)
- LRU cache for efficient memory management
- Asynchronous loading with proper state management
- Graceful error handling with fallback to alt text

## Architecture

### Components

#### 1. Image Loader (`internal/image/loader.go`)

The `Loader` handles fetching and decoding images from various sources.

**Features:**
- Asynchronous loading to avoid blocking the UI
- Synchronous loading option for cases where immediate results are needed
- Automatic detection of URL vs file path
- Deduplication of concurrent requests for the same image
- Integration with the cache system

**Load States:**
- `StateLoading`: Image is being fetched and decoded
- `StateLoaded`: Image successfully loaded
- `StateError`: Loading failed (with error details)

**Usage:**
```go
loader := image.NewLoader(100) // Cache up to 100 images

// Async load (returns immediately with StateLoading)
imageData, err := loader.Load("https://example.com/image.png")

// Sync load (waits for completion)
imageData, err := loader.LoadSync("/path/to/local/image.png")
```

#### 2. Image Cache (`internal/image/cache.go`)

The `Cache` implements an LRU (Least Recently Used) eviction policy.

**Features:**
- Thread-safe operations with mutex locks
- Automatic eviction when capacity is reached
- Configurable capacity
- O(1) get and put operations

**Usage:**
```go
cache := image.NewCache(50) // Capacity of 50 images

cache.Put("key", imageData)
data := cache.Get("key")
cache.Clear()
```

#### 3. Renderer Integration

The renderer automatically uses the image loader for `<img>` tags:

**Canvas Renderer (`internal/renderer/canvas.go`):**
- Receives image loader instance during initialization
- Attempts to load images when rendering img elements
- Shows different placeholders based on load state:
  - Loading: Gray rectangle with "Loading Image" text
  - Loaded: Actual image with optional alt text
  - Error: "Image Load Failed" text with alt text

**Display List (`internal/renderer/display_list.go`):**
- Generates `PaintImage` commands for img elements
- Includes source URL and alt text for rendering

## Supported Image Formats

The system supports the following image formats through Go's standard library:
- **PNG**: `image/png`
- **JPEG**: `image/jpeg`
- **GIF**: `image/gif`
- **WebP**: `golang.org/x/image/webp`

Additional formats can be added by importing the appropriate decoders.

## Cache Eviction Policy

The cache uses an LRU (Least Recently Used) eviction policy:

1. When an image is accessed (Get), it's moved to the front of the list
2. When a new image is added and capacity is reached, the least recently used image (at the back of the list) is removed
3. The cache maintains O(1) complexity for both get and put operations using a combination of:
   - HashMap for fast key lookup
   - Doubly-linked list for LRU ordering

## Error Handling

The system handles errors gracefully at multiple levels:

1. **Network Errors**: Caught during HTTP fetch, cached with StateError
2. **File Errors**: Caught when opening local files, cached with StateError
3. **Decode Errors**: Caught when decoding image data, cached with StateError
4. **Missing Source**: Img tags without src attribute show alt text only

Failed loads are cached (with error state) to avoid repeated failed requests.

## Performance Considerations

1. **Asynchronous Loading**: Images load in the background without blocking the UI
2. **Cache Reuse**: Once loaded, images are served from cache
3. **Request Deduplication**: Concurrent requests for the same image wait for a single load
4. **Memory Management**: LRU cache automatically evicts old images to stay within capacity
5. **Viewport Optimization**: Only images in or near the viewport need to be loaded

## Testing

Comprehensive tests are provided:

**Unit Tests:**
- `cache_test.go`: Tests cache operations, eviction, and LRU ordering
- `loader_test.go`: Tests loading from URLs and files, error handling, caching

**Integration Tests:**
- `image_integration_test.go`: Tests image rendering in the renderer
- Tests for missing images, images without source, and cache eviction

## Future Enhancements

Potential improvements for the future:
- Image resizing/scaling for responsive layouts
- Progressive image loading for large images
- Prefetching images that are about to enter the viewport
- Image compression for cache efficiency
- Support for SVG images
- Image lazy loading attributes
- Background image support (CSS)

## Usage Examples

### Basic Image Rendering

```html
<img src="https://example.com/image.png" alt="Description">
```

### Local Image

```html
<img src="/path/to/local/image.jpg" alt="Local Image">
```

### Image Without Source

```html
<img alt="No Image">
```
Shows: `[Image: No Image]`

### Failed Image Load

If an image fails to load:
```
[Image Load Failed: Description]
```

## API Reference

### Loader

```go
type Loader struct { ... }

// Create new loader with cache size
func NewLoader(cacheSize int) *Loader

// Load image asynchronously
func (l *Loader) Load(source string) (*ImageData, error)

// Load image synchronously
func (l *Loader) LoadSync(source string) (*ImageData, error)

// Get cache instance
func (l *Loader) GetCache() *Cache
```

### Cache

```go
type Cache struct { ... }

// Create new cache with capacity
func NewCache(capacity int) *Cache

// Get image from cache
func (c *Cache) Get(key string) *ImageData

// Put image in cache
func (c *Cache) Put(key string, value *ImageData)

// Clear all cached images
func (c *Cache) Clear()

// Get current cache size
func (c *Cache) Len() int

// Update cache capacity
func (c *Cache) SetCapacity(capacity int)
```

### ImageData

```go
type ImageData struct {
    Image  image.Image  // Decoded image
    Width  int          // Image width in pixels
    Height int          // Image height in pixels
    Format string       // Image format (png, jpeg, gif, webp)
    State  LoadState    // Current load state
    Error  error        // Error if State is StateError
}
```
