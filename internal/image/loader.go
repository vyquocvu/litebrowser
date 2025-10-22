package image

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	_ "golang.org/x/image/webp"
)

// LoadState represents the state of an image load operation
type LoadState int

const (
	// StateLoading indicates the image is being loaded
	StateLoading LoadState = iota
	// StateLoaded indicates the image was successfully loaded
	StateLoaded
	// StateError indicates an error occurred during loading
	StateError
)

// ImageData represents a loaded image with metadata
type ImageData struct {
	Image  image.Image
	Width  int
	Height int
	Format string
	State  LoadState
	Error  error
}

// Loader handles loading images from various sources
type Loader struct {
	httpClient *http.Client
	cache      *Cache
	mu         sync.RWMutex
	// Track in-progress loads to avoid duplicate requests
	inProgress map[string]*sync.WaitGroup
}

// NewLoader creates a new image loader with a cache
func NewLoader(cacheSize int) *Loader {
	return &Loader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:      NewCache(cacheSize),
		inProgress: make(map[string]*sync.WaitGroup),
	}
}

// Load loads an image from a URL or file path
// Returns cached image if available, otherwise loads asynchronously
func (l *Loader) Load(source string) (*ImageData, error) {
	// Check cache first
	if cached := l.cache.Get(source); cached != nil {
		return cached, nil
	}

	// Check if already loading this image
	l.mu.Lock()
	if wg, exists := l.inProgress[source]; exists {
		l.mu.Unlock()
		// Wait for the in-progress load to complete
		wg.Wait()
		// Try cache again after waiting
		if cached := l.cache.Get(source); cached != nil {
			return cached, nil
		}
		// If still not in cache, it failed - return loading state
		return &ImageData{State: StateLoading}, nil
	}

	// Mark as in-progress
	wg := &sync.WaitGroup{}
	wg.Add(1)
	l.inProgress[source] = wg
	l.mu.Unlock()

	// Return loading state immediately and load in background
	go l.loadAsync(source, wg)

	return &ImageData{State: StateLoading}, nil
}

// LoadSync loads an image synchronously
func (l *Loader) LoadSync(source string) (*ImageData, error) {
	// Check cache first
	if cached := l.cache.Get(source); cached != nil {
		return cached, nil
	}

	// Load the image
	data, err := l.loadImage(source)
	if err != nil {
		data = &ImageData{
			State: StateError,
			Error: err,
		}
	}

	// Cache the result (even errors)
	l.cache.Put(source, data)

	return data, err
}

// loadAsync loads an image asynchronously
func (l *Loader) loadAsync(source string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		l.mu.Lock()
		delete(l.inProgress, source)
		l.mu.Unlock()
	}()

	data, err := l.loadImage(source)
	if err != nil {
		data = &ImageData{
			State: StateError,
			Error: err,
		}
	}

	// Cache the result
	l.cache.Put(source, data)
}

// loadImage loads an image from a source (URL or file path)
func (l *Loader) loadImage(source string) (*ImageData, error) {
	// Determine if it's a URL or file path
	if isURL(source) {
		return l.loadFromURL(source)
	}
	return l.loadFromFile(source)
}

// loadFromURL loads an image from a remote URL
func (l *Loader) loadFromURL(url string) (*ImageData, error) {
	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return l.decodeImage(bytes.NewReader(data))
}

// loadFromFile loads an image from a local file
func (l *Loader) loadFromFile(path string) (*ImageData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	return l.decodeImage(file)
}

// decodeImage decodes an image from a reader
func (l *Loader) decodeImage(r io.Reader) (*ImageData, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	return &ImageData{
		Image:  img,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
		State:  StateLoaded,
	}, nil
}

// GetCache returns the cache instance
func (l *Loader) GetCache() *Cache {
	return l.cache
}

// isURL checks if a string is a URL
func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}
