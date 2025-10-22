package renderer

import (
	"image/color"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	
	imageloader "github.com/vyquocvu/litebrowser/internal/image"
)

// CanvasRenderer renders a render tree onto a Fyne canvas
type CanvasRenderer struct {
	canvasWidth  float32
	canvasHeight float32
	defaultSize  float32

	// Viewport for optimized rendering
	viewportY      float32
	viewportHeight float32

	// Cached display list for performance
	cachedDisplayList *DisplayList
	cachedLayoutRoot  *LayoutBox
	cachedRenderRoot  *RenderNode
	
	// fontMetrics provides accurate text measurement
	fontMetrics *FontMetrics
	
	// Navigation callback for link clicks
	onNavigate NavigationCallback
	
	// Current page URL for resolving relative links
	baseURL string
	
	// Image loader for loading and caching images
	imageLoader *imageloader.Loader
}

// NewCanvasRenderer creates a new canvas renderer
func NewCanvasRenderer(width, height float32) *CanvasRenderer {
	defaultSize := float32(16.0)
	return &CanvasRenderer{
		canvasWidth:    width,
		canvasHeight:   height,
		defaultSize:    defaultSize,
		viewportY:      0,
		viewportHeight: height,
		fontMetrics:    NewFontMetrics(defaultSize),
	}
}

// SetViewport sets the current viewport for optimized rendering
func (cr *CanvasRenderer) SetViewport(y, height float32) {
	cr.viewportY = y
	cr.viewportHeight = height
}

// SetNavigationCallback sets the navigation callback for link clicks
func (cr *CanvasRenderer) SetNavigationCallback(callback NavigationCallback, baseURL string) {
	cr.onNavigate = callback
	cr.baseURL = baseURL
}

// isInViewport checks if a box intersects with the current viewport
func (cr *CanvasRenderer) isInViewport(box Rect) bool {
	// Add buffer zone above and below viewport for smoother scrolling
	bufferZone := cr.viewportHeight * 0.5
	viewportTop := cr.viewportY - bufferZone
	viewportBottom := cr.viewportY + cr.viewportHeight + bufferZone

	boxBottom := box.Y + box.Height

	// Check if box intersects with viewport
	return boxBottom >= viewportTop && box.Y <= viewportBottom
}

// Render renders the render tree and returns a Fyne container
func (cr *CanvasRenderer) Render(root *RenderNode) fyne.CanvasObject {
	if root == nil {
		return container.NewVBox()
	}

	objects := make([]fyne.CanvasObject, 0)
	cr.renderNode(root, &objects)

	return container.NewVBox(objects...)
}

// renderNode renders a single node and its children
func (cr *CanvasRenderer) renderNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	if node == nil {
		return
	}

	if node.Type == NodeTypeText {
		cr.renderTextNode(node, objects)
	} else if node.Type == NodeTypeElement {
		cr.renderElementNode(node, objects)
	}
}

// renderTextNode renders a text node
func (cr *CanvasRenderer) renderTextNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := strings.TrimSpace(node.Text)
	if text == "" {
		return
	}

	// Create text widget
	textWidget := widget.NewLabel(text)
	textWidget.Wrapping = fyne.TextWrapWord

	// Get text style from parent if available
	if node.Parent != nil {
		textWidget.TextStyle = cr.fontMetrics.GetTextStyle(node.Parent.TagName)
	}

	*objects = append(*objects, textWidget)
}

// renderElementNode renders an element node
func (cr *CanvasRenderer) renderElementNode(node *RenderNode, objects *[]fyne.CanvasObject) {
	switch node.TagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		cr.renderHeading(node, objects)
	case "p":
		cr.renderParagraph(node, objects)
	case "div":
		cr.renderDiv(node, objects)
	case "a":
		cr.renderLink(node, objects)
	case "ul", "ol":
		cr.renderList(node, objects)
	case "li":
		cr.renderListItem(node, objects)
	case "img":
		cr.renderImage(node, objects)
	case "br":
		// Add a spacer for line break
		*objects = append(*objects, widget.NewLabel(""))
	case "span", "strong", "em", "b", "i":
		// Inline elements - render children
		for _, child := range node.Children {
			cr.renderNode(child, objects)
		}
	default:
		// Generic element - just render children
		for _, child := range node.Children {
			cr.renderNode(child, objects)
		}
	}
}

// renderHeading renders heading elements (h1-h6)
func (cr *CanvasRenderer) renderHeading(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Different sizes for different heading levels
	// Note: Fyne doesn't support arbitrary font sizes directly,
	// so we use TextStyle to make headings bold

	*objects = append(*objects, label)
}

// renderParagraph renders paragraph elements
func (cr *CanvasRenderer) renderParagraph(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	*objects = append(*objects, label)
}

// renderDiv renders div elements
func (cr *CanvasRenderer) renderDiv(node *RenderNode, objects *[]fyne.CanvasObject) {
	// Render children
	for _, child := range node.Children {
		cr.renderNode(child, objects)
	}
}

// renderLink renders anchor (link) elements
func (cr *CanvasRenderer) renderLink(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	href, hasHref := node.GetAttribute("href")

	if text == "" {
		return
	}

	if hasHref && href != "" {
		// Resolve URL (absolute or relative)
		resolvedURL := cr.resolveURL(href)
		
		// Note: Link target attribute (_blank, _self, etc.) is available via node.GetAttribute("target")
		// but not currently implemented as the browser doesn't support tabs yet.
		// This is planned for Phase 1 UI Improvements (see ROADMAP.md).
		
		// Parse URL to create a proper Fyne URL object
		parsedURL, err := url.Parse(resolvedURL)
		if err != nil {
			// If URL parsing fails, display as text
			label := widget.NewLabel(text)
			label.Wrapping = fyne.TextWrapWord
			*objects = append(*objects, label)
			return
		}
		
		// Create a clickable hyperlink widget
		link := widget.NewHyperlink(text, parsedURL)
		link.Wrapping = fyne.TextWrapWord
		
		// Override the default tap handler to use our navigation callback
		if cr.onNavigate != nil {
			// Create a custom tappable widget
			tappableLink := newTappableHyperlink(text, resolvedURL, cr.onNavigate)
			*objects = append(*objects, tappableLink)
		} else {
			// Fallback to default hyperlink behavior
			*objects = append(*objects, link)
		}
	} else {
		// No href, just display as text
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		*objects = append(*objects, label)
	}
}

// resolveURL resolves a relative or absolute URL against the base URL
func (cr *CanvasRenderer) resolveURL(href string) string {
	// If href is already absolute, return as-is
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	
	// If no base URL, return href as-is
	if cr.baseURL == "" {
		return href
	}
	
	// Parse base URL
	baseURL, err := url.Parse(cr.baseURL)
	if err != nil {
		return href
	}
	
	// Parse relative href
	relURL, err := url.Parse(href)
	if err != nil {
		return href
	}
	
	// Resolve relative URL against base
	resolved := baseURL.ResolveReference(relURL)
	return resolved.String()
}

// TappableHyperlink is a custom hyperlink widget that can trigger navigation callbacks.
// It extends widget.Hyperlink, inheriting keyboard navigation support (Tab focus, Enter activation).
type TappableHyperlink struct {
	widget.Hyperlink
	url        string
	onNavigate NavigationCallback
}

// newTappableHyperlink creates a new tappable hyperlink
func newTappableHyperlink(text, urlStr string, onNavigate NavigationCallback) *TappableHyperlink {
	parsedURL := urlParse(urlStr)
	link := &TappableHyperlink{
		url:        urlStr,
		onNavigate: onNavigate,
	}
	link.ExtendBaseWidget(link)
	link.Text = text
	link.URL = parsedURL
	link.Wrapping = fyne.TextWrapWord
	return link
}

// Tapped handles tap events on the hyperlink
func (t *TappableHyperlink) Tapped(_ *fyne.PointEvent) {
	if t.onNavigate != nil {
		t.onNavigate(t.url)
	}
}

// urlParse is a helper that returns nil on parse error
func urlParse(urlStr string) *url.URL {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return parsed
}

// renderList renders ul/ol elements
func (cr *CanvasRenderer) renderList(node *RenderNode, objects *[]fyne.CanvasObject) {
	// Render list items
	for _, child := range node.Children {
		if child.TagName == "li" {
			cr.renderListItem(child, objects)
		}
	}
}

// renderListItem renders li elements
func (cr *CanvasRenderer) renderListItem(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	// Add bullet point
	label := widget.NewLabel("• " + text)
	label.Wrapping = fyne.TextWrapWord

	*objects = append(*objects, label)
}

// renderImage renders img elements
func (cr *CanvasRenderer) renderImage(node *RenderNode, objects *[]fyne.CanvasObject) {
	alt, hasAlt := node.GetAttribute("alt")
	src, hasSrc := node.GetAttribute("src")

	if !hasSrc || src == "" {
		// No source - show alt text or placeholder
		displayText := "[Image"
		if hasAlt {
			displayText += ": " + alt
		}
		displayText += "]"
		label := widget.NewLabel(displayText)
		label.Wrapping = fyne.TextWrapWord
		*objects = append(*objects, label)
		return
	}

	// Resolve relative URLs
	resolvedSrc := cr.resolveURL(src)

	// Try to load the image if loader is available
	if cr.imageLoader != nil {
		imageData, err := cr.imageLoader.Load(resolvedSrc)
		
		if err == nil && imageData != nil {
			switch imageData.State {
			case imageloader.StateLoaded:
				// Image loaded successfully - render it
				img := canvas.NewImageFromImage(imageData.Image)
				img.FillMode = canvas.ImageFillOriginal
				img.SetMinSize(fyne.NewSize(float32(imageData.Width), float32(imageData.Height)))
				
				// Add alt text below the image if available
				if hasAlt && alt != "" {
					altLabel := widget.NewLabel(alt)
					altLabel.Wrapping = fyne.TextWrapWord
					*objects = append(*objects, container.NewVBox(img, altLabel))
				} else {
					*objects = append(*objects, img)
				}
				return
				
			case imageloader.StateError:
				// Image failed to load - show error with alt text
				displayText := "[Image Load Failed"
				if hasAlt {
					displayText += ": " + alt
				}
				displayText += "]"
				label := widget.NewLabel(displayText)
				label.Wrapping = fyne.TextWrapWord
				*objects = append(*objects, label)
				return
				
			case imageloader.StateLoading:
				// Image is loading - show loading placeholder
				displayText := "[Loading Image"
				if hasAlt {
					displayText += ": " + alt
				}
				displayText += "]"
				label := widget.NewLabel(displayText)
				label.Wrapping = fyne.TextWrapWord
				
				// Show a gray rectangle as loading indicator
				rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
				rect.SetMinSize(fyne.NewSize(100, 100))
				
				*objects = append(*objects, container.NewVBox(rect, label))
				return
			}
		}
	}

	// Fallback: Show placeholder if loader is not available or something went wrong
	displayText := "[Image: " + src
	if hasAlt {
		displayText += " - " + alt
	}
	displayText += "]"

	label := widget.NewLabel(displayText)
	label.Wrapping = fyne.TextWrapWord

	rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	rect.SetMinSize(fyne.NewSize(100, 100))

	*objects = append(*objects, container.NewVBox(rect, label))
}

// extractText extracts all text content from a node and its children
func (cr *CanvasRenderer) extractText(node *RenderNode) string {
	var text strings.Builder
	cr.extractTextRecursive(node, &text)
	return strings.TrimSpace(text.String())
}

// extractTextRecursive recursively extracts text from a node tree
func (cr *CanvasRenderer) extractTextRecursive(node *RenderNode, builder *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == NodeTypeText {
		text := strings.TrimSpace(node.Text)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(text)
		}
	}

	for _, child := range node.Children {
		cr.extractTextRecursive(child, builder)
	}
}

// getFontSize returns font size for an element type (delegates to fontMetrics)
func (cr *CanvasRenderer) getFontSize(tagName string) float32 {
	return cr.fontMetrics.GetFontSize(tagName)
}

// getTextStyle returns text style for an element type (delegates to fontMetrics)
func (cr *CanvasRenderer) getTextStyle(tagName string) fyne.TextStyle {
	return cr.fontMetrics.GetTextStyle(tagName)
}

// RenderWithViewport renders the render tree with viewport culling for better performance
func (cr *CanvasRenderer) RenderWithViewport(root *RenderNode, layoutRoot *LayoutBox) fyne.CanvasObject {
	if root == nil || layoutRoot == nil {
		return container.NewVBox()
	}

	// Build or reuse display list
	var displayList *DisplayList
	if cr.cachedDisplayList != nil && cr.cachedRenderRoot == root && cr.cachedLayoutRoot == layoutRoot {
		// Reuse cached display list
		displayList = cr.cachedDisplayList
	} else {
		// Build new display list
		dlb := NewDisplayListBuilder()
		displayList = dlb.Build(layoutRoot, root)

		// Cache for next time
		cr.cachedDisplayList = displayList
		cr.cachedRenderRoot = root
		cr.cachedLayoutRoot = layoutRoot
	}

	// Filter commands based on viewport
	objects := make([]fyne.CanvasObject, 0)
	for _, cmd := range displayList.Commands {
		if cr.isInViewport(cmd.Box) {
			cr.renderCommand(cmd, &objects)
		}
	}

	if len(objects) == 0 {
		return container.NewVBox()
	}

	return container.NewVBox(objects...)
}

// renderCommand renders a single paint command to canvas objects
func (cr *CanvasRenderer) renderCommand(cmd *PaintCommand, objects *[]fyne.CanvasObject) {
	switch cmd.Type {
	case PaintText:
		if strings.TrimSpace(cmd.Text) == "" {
			return
		}

		label := widget.NewLabel(cmd.Text)
		label.Wrapping = fyne.TextWrapWord

		if cmd.Bold && cmd.Italic {
			label.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
		} else if cmd.Bold {
			label.TextStyle = fyne.TextStyle{Bold: true}
		} else if cmd.Italic {
			label.TextStyle = fyne.TextStyle{Italic: true}
		}

		*objects = append(*objects, label)

	case PaintRect:
		rect := canvas.NewRectangle(cmd.FillColor)
		rect.SetMinSize(fyne.NewSize(cmd.Box.Width, cmd.Box.Height))
		*objects = append(*objects, rect)

	case PaintImage:
		// Try to load and render the actual image if loader is available
		if cr.imageLoader != nil && cmd.ImageSrc != "" {
			resolvedSrc := cr.resolveURL(cmd.ImageSrc)
			imageData, err := cr.imageLoader.Load(resolvedSrc)
			
			if err == nil && imageData != nil {
				switch imageData.State {
				case imageloader.StateLoaded:
					// Image loaded successfully - render it
					img := canvas.NewImageFromImage(imageData.Image)
					img.FillMode = canvas.ImageFillOriginal
					img.SetMinSize(fyne.NewSize(float32(imageData.Width), float32(imageData.Height)))
					
					// Add alt text below the image if available
					if cmd.ImageAlt != "" {
						altLabel := widget.NewLabel(cmd.ImageAlt)
						altLabel.Wrapping = fyne.TextWrapWord
						*objects = append(*objects, container.NewVBox(img, altLabel))
					} else {
						*objects = append(*objects, img)
					}
					return
					
				case imageloader.StateError:
					// Image failed to load - show error with alt text
					displayText := "[Image Load Failed"
					if cmd.ImageAlt != "" {
						displayText += ": " + cmd.ImageAlt
					}
					displayText += "]"
					label := widget.NewLabel(displayText)
					label.Wrapping = fyne.TextWrapWord
					*objects = append(*objects, label)
					return
					
				case imageloader.StateLoading:
					// Image is loading - show loading placeholder
					displayText := "[Loading Image"
					if cmd.ImageAlt != "" {
						displayText += ": " + cmd.ImageAlt
					}
					displayText += "]"
					label := widget.NewLabel(displayText)
					label.Wrapping = fyne.TextWrapWord
					
					rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
					rect.SetMinSize(fyne.NewSize(100, 100))
					
					*objects = append(*objects, container.NewVBox(rect, label))
					return
				}
			}
		}
		
		// Fallback: Render image placeholder
		displayText := "[Image"
		if cmd.ImageSrc != "" {
			displayText += ": " + cmd.ImageSrc
		}
		if cmd.ImageAlt != "" {
			displayText += " - " + cmd.ImageAlt
		}
		displayText += "]"

		label := widget.NewLabel(displayText)
		label.Wrapping = fyne.TextWrapWord

		rect := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
		rect.SetMinSize(fyne.NewSize(100, 100))

		*objects = append(*objects, container.NewVBox(rect, label))
		
	case PaintLink:
		// Render clickable link
		if cmd.LinkText == "" {
			return
		}
		
		// Resolve URL (absolute or relative)
		resolvedURL := cr.resolveURL(cmd.LinkURL)
		
		// Create a clickable hyperlink widget
		if cr.onNavigate != nil {
			// Create a custom tappable widget
			tappableLink := newTappableHyperlink(cmd.LinkText, resolvedURL, cr.onNavigate)
			*objects = append(*objects, tappableLink)
		} else {
			// Fallback to default hyperlink behavior
			parsedURL, err := url.Parse(resolvedURL)
			if err == nil {
				link := widget.NewHyperlink(cmd.LinkText, parsedURL)
				link.Wrapping = fyne.TextWrapWord
				*objects = append(*objects, link)
			} else {
				// If URL parsing fails, display as text
				label := widget.NewLabel(cmd.LinkText)
				label.Wrapping = fyne.TextWrapWord
				*objects = append(*objects, label)
			}
		}
	}
}

// ClearCache clears the cached display list to force re-rendering
func (cr *CanvasRenderer) ClearCache() {
	cr.cachedDisplayList = nil
	cr.cachedLayoutRoot = nil
	cr.cachedRenderRoot = nil
}
