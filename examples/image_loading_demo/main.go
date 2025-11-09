package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	imageloader "github.com/vyquocvu/goosie/internal/image"
)

// This example demonstrates the image loading and caching capabilities
func main() {
	fmt.Println("Image Loading Demo")
	fmt.Println("==================")

	// Create a temporary directory for test images
	tmpDir, err := os.MkdirTemp("", "image-demo-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple test image
	fmt.Println("Creating test image...")
	testImagePath := filepath.Join(tmpDir, "test.png")
	if err := createTestImage(testImagePath, 100, 100); err != nil {
		fmt.Printf("Failed to create test image: %v\n", err)
		return
	}
	fmt.Printf("Created test image at: %s\n\n", testImagePath)

	// Example 1: Create an image loader with cache
	fmt.Println("Example 1: Creating Image Loader")
	fmt.Println("---------------------------------")
	loader := imageloader.NewLoader(50) // Cache up to 50 images
	fmt.Printf("Created loader with cache capacity: %d\n\n", 50)

	// Example 2: Load an image synchronously
	fmt.Println("Example 2: Loading Image Synchronously")
	fmt.Println("---------------------------------------")
	imageData, err := loader.Load(testImagePath)
	if err != nil {
		fmt.Printf("Error loading image: %v\n", err)
	} else {
		fmt.Printf("Image loaded successfully!\n")
		fmt.Printf("  Width: %d pixels\n", imageData.Width)
		fmt.Printf("  Height: %d pixels\n", imageData.Height)
		fmt.Printf("  Format: %s\n", imageData.Format)
		fmt.Printf("  State: %v\n\n", imageData.State)
	}

	// Example 3: Check cache
	fmt.Println("Example 3: Cache Status")
	fmt.Println("-----------------------")
	fmt.Printf("Cache size: %d\n", loader.GetCache().Len())
	cached := loader.GetCache().Get(testImagePath)
	if cached != nil {
		fmt.Printf("Image is cached: %dx%d %s\n\n", cached.Width, cached.Height, cached.Format)
	}

	// Example 4: Load from cache
	fmt.Println("Example 4: Loading from Cache")
	fmt.Println("------------------------------")
	_, err = loader.Load(testImagePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Image loaded from cache!\n")
		fmt.Printf("Cache still has %d item(s)\n\n", loader.GetCache().Len())
	}

	// Example 5: Error handling
	fmt.Println("Example 5: Handling Missing Images")
	fmt.Println("-----------------------------------")
	missingData, err := loader.Load("/nonexistent/image.png")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		if missingData != nil && missingData.State == 2 { // StateError
			fmt.Printf("Image state correctly set to Error\n")
		}
	}
	fmt.Println()

	// Example 6: Cache eviction
	fmt.Println("Example 6: Cache Eviction Demo")
	fmt.Println("-------------------------------")
	smallLoader := imageloader.NewLoader(2) // Small cache

	// Create 3 test images
	for i := 1; i <= 3; i++ {
		imagePath := filepath.Join(tmpDir, fmt.Sprintf("image%d.png", i))
		if err := createTestImage(imagePath, 50, 50); err != nil {
			fmt.Printf("Failed to create image %d: %v\n", i, err)
			continue
		}
		smallLoader.Load(imagePath)
		fmt.Printf("Loaded image %d, cache size: %d\n", i, smallLoader.GetCache().Len())
	}
	fmt.Printf("\nFinal cache size (should be 2 due to eviction): %d\n", smallLoader.GetCache().Len())

	fmt.Println("\n✓ Image loading demo complete!")
}

// createTestImage creates a simple colored test image
func createTestImage(path string, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
