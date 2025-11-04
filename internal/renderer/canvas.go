package renderer

import (
	"image/color"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	imageloader "github.com/vyquocvu/goosie/internal/image"
)

// CanvasRenderer renders a render tree onto a Fyne canvas
type CanvasRenderer struct {
	canvasWidth  float32
	canvasHeight float32
	defaultSize  float32
	window       fyne.Window

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
	imageLoader imageloader.Loader

	// OnRefresh is a test hook to signal when a refresh is triggered.
	OnRefresh func()
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

// SetWindow sets the Fyne window for the renderer
func (cr *CanvasRenderer) SetWindow(w fyne.Window) {
	cr.window = w
	if cr.imageLoader != nil {
		cr.imageLoader.SetOnLoadCallback(cr.onImageLoaded)
	}
}

func (cr *CanvasRenderer) onImageLoaded(source string) {
	if cr.window == nil {
		return
	}

	// Use fyne.Do to safely update the UI from any thread
	fyne.Do(func() {
		cr.ClearCache()
		cr.window.Canvas().Refresh(cr.window.Content())
		
		if cr.OnRefresh != nil {
			cr.OnRefresh()
		}
	})
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
	case "input":
		cr.renderInput(node, objects)
	case "button":
		cr.renderButton(node, objects)
	case "textarea":
		cr.renderTextarea(node, objects)
	case "table":
		cr.renderTable(node, objects)
	case "tbody", "thead", "tfoot":
		// These are handled by renderTable, but if rendered independently,
		// treat them as containers
		for _, child := range node.Children {
			cr.renderNode(child, objects)
		}
	case "tr":
		// Handled by renderTable
	case "td", "th":
		// Handled by renderTable
	case "br":
		// Add a spacer for line break
		*objects = append(*objects, widget.NewLabel(""))
	case "code":
		cr.renderCode(node, objects)
	case "pre":
		cr.renderPre(node, objects)
	case "blockquote":
		cr.renderBlockquote(node, objects)
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

	// Apply CSS styles if present
	styledObj := cr.applyStylesToLabel(node, text)
	
	// If it's a standard label (no CSS), apply heading styles
	if label, ok := styledObj.(*widget.Label); ok {
		label.TextStyle = fyne.TextStyle{Bold: true}
	}

	*objects = append(*objects, styledObj)
}

// renderParagraph renders paragraph elements
func (cr *CanvasRenderer) renderParagraph(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	// Apply CSS styles if present
	styledObj := cr.applyStylesToLabel(node, text)
	*objects = append(*objects, styledObj)
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

// renderCode renders code elements with monospace styling
func (cr *CanvasRenderer) renderCode(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle = fyne.TextStyle{Monospace: true}

	*objects = append(*objects, label)
}

// renderPre renders pre elements with monospace styling and preserved whitespace
func (cr *CanvasRenderer) renderPre(node *RenderNode, objects *[]fyne.CanvasObject) {
	// For pre elements, we want to preserve whitespace and newlines
	// Extract text without trimming
	text := cr.extractTextPreserveWhitespace(node)
	if text == "" {
		return
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapOff // Pre elements typically don't wrap
	label.TextStyle = fyne.TextStyle{Monospace: true}

	*objects = append(*objects, label)
}

// renderBlockquote renders blockquote elements
func (cr *CanvasRenderer) renderBlockquote(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	if text == "" {
		return
	}

	// Add visual indication of quote (e.g., with prefix)
	label := widget.NewLabel("❝ " + text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle = fyne.TextStyle{Italic: true}

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

// extractTextPreserveWhitespace extracts text content while preserving whitespace and newlines
// This is used for <pre> elements where whitespace formatting is significant
func (cr *CanvasRenderer) extractTextPreserveWhitespace(node *RenderNode) string {
	var text strings.Builder
	cr.extractTextPreserveWhitespaceRecursive(node, &text)
	return text.String()
}

// extractTextPreserveWhitespaceRecursive recursively extracts text without trimming whitespace
func (cr *CanvasRenderer) extractTextPreserveWhitespaceRecursive(node *RenderNode, builder *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == NodeTypeText {
		// Don't trim whitespace for pre elements
		builder.WriteString(node.Text)
	}

	for _, child := range node.Children {
		cr.extractTextPreserveWhitespaceRecursive(child, builder)
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

		// Check if the node has CSS styles
		hasCustomStyles := cmd.Node != nil && cmd.Node.ComputedStyle != nil && (
			cmd.Node.ComputedStyle.Color != nil ||
			cmd.Node.ComputedStyle.FontSize > 0 ||
			cmd.Node.ComputedStyle.FontWeight == "bold")

		if hasCustomStyles {
			// Create a canvas.Text object with CSS styles
			textObj := canvas.NewText(cmd.Text, color.Black)
			textObj.TextSize = cr.defaultSize

			style := cmd.Node.ComputedStyle

			if style.Color != nil {
				textObj.Color = style.Color
			}

			if style.FontSize > 0 {
				textObj.TextSize = style.FontSize
			}

			// Apply text style
			textStyle := fyne.TextStyle{}
			if style.FontWeight == "bold" || cmd.Bold {
				textStyle.Bold = true
			}
			if cmd.Italic {
				textStyle.Italic = true
			}
			textObj.TextStyle = textStyle

			*objects = append(*objects, textObj)
		} else {
			// Use standard label widget
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
		}

	case PaintRect:
		rect := canvas.NewRectangle(cmd.FillColor)
		rect.SetMinSize(fyne.NewSize(cmd.Box.Width, cmd.Box.Height))
		*objects = append(*objects, rect)

	case PaintImage:
		// Try to load and render the actual image if loader is available
		if cr.imageLoader != nil && cmd.Node.ImageData != nil {
			imageData := cmd.Node.ImageData

			if imageData != nil {
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

func (cr *CanvasRenderer) renderInput(node *RenderNode, objects *[]fyne.CanvasObject) {
	entry := widget.NewEntry()
	if placeholder, ok := node.GetAttribute("placeholder"); ok {
		entry.SetPlaceHolder(placeholder)
	}
	*objects = append(*objects, entry)
}

func (cr *CanvasRenderer) renderTable(node *RenderNode, objects *[]fyne.CanvasObject) {
	data := [][]string{}
	var maxCols int

	// Helper function to extract rows from a node (handles tbody, thead, tfoot)
	var extractRows func(*RenderNode)
	extractRows = func(n *RenderNode) {
		for _, child := range n.Children {
			if child.TagName == "tr" {
				row := []string{}
				for _, td := range child.Children {
					if td.TagName == "td" || td.TagName == "th" {
						row = append(row, cr.extractText(td))
					}
				}
				if len(row) > maxCols {
					maxCols = len(row)
				}
				data = append(data, row)
			} else if child.TagName == "tbody" || child.TagName == "thead" || child.TagName == "tfoot" {
				// Recursively process tbody, thead, tfoot
				extractRows(child)
			}
		}
	}

	extractRows(node)

	if len(data) == 0 || maxCols == 0 {
		return
	}

	table := widget.NewTable(
		func() (int, int) {
			return len(data), maxCols
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row < len(data) && i.Col < len(data[i.Row]) {
				o.(*widget.Label).SetText(data[i.Row][i.Col])
			}
		},
	)

	for i := 0; i < maxCols; i++ {
		table.SetColumnWidth(i, 100)
	}

	*objects = append(*objects, table)
}

func (cr *CanvasRenderer) renderButton(node *RenderNode, objects *[]fyne.CanvasObject) {
	text := cr.extractText(node)
	button := widget.NewButton(text, func() {})
	*objects = append(*objects, button)
}

func (cr *CanvasRenderer) renderTextarea(node *RenderNode, objects *[]fyne.CanvasObject) {
	entry := widget.NewMultiLineEntry()
	if placeholder, ok := node.GetAttribute("placeholder"); ok {
		entry.SetPlaceHolder(placeholder)
	}
	*objects = append(*objects, entry)
}

// applyStylesToLabel applies CSS styles from ComputedStyle to a label widget
// Since Fyne's standard Label widget doesn't support custom colors or font sizes,
// this function creates a styled canvas.Text object when custom styles are present
func (cr *CanvasRenderer) applyStylesToLabel(node *RenderNode, text string) fyne.CanvasObject {
	// Check if the node has custom styles
	hasCustomStyles := node.ComputedStyle != nil && (
		node.ComputedStyle.Color != nil ||
		node.ComputedStyle.FontSize > 0 ||
		node.ComputedStyle.FontWeight == "bold")

	if !hasCustomStyles {
		// No custom styles, use standard label
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		
		// Apply tag-based styles (bold, italic, etc.)
		if node.Parent != nil {
			label.TextStyle = cr.fontMetrics.GetTextStyle(node.Parent.TagName)
		}
		
		return label
	}

	// Create a styled canvas.Text object
	textObj := canvas.NewText(text, color.Black)
	textObj.TextSize = cr.defaultSize
	
	// Apply computed styles
	style := node.ComputedStyle
	
	if style.Color != nil {
		textObj.Color = style.Color
	}
	
	if style.FontSize > 0 {
		textObj.TextSize = style.FontSize
	}
	
	if style.FontWeight == "bold" {
		textObj.TextStyle = fyne.TextStyle{Bold: true}
	}
	
	// Fyne's canvas.Text doesn't support wrapping directly,
	// so we need to wrap it in a container with RichText for wrapping
	// For simplicity, we'll use the text object as-is for now
	return textObj
}

// getEffectiveFontSize returns the font size for a node, considering CSS and default sizes
func (cr *CanvasRenderer) getEffectiveFontSize(node *RenderNode) float32 {
	if node.ComputedStyle != nil && node.ComputedStyle.FontSize > 0 {
		return node.ComputedStyle.FontSize
	}
	if node.Parent != nil {
		return cr.fontMetrics.GetFontSize(node.Parent.TagName)
	}
	return cr.defaultSize
}

// getEffectiveColor returns the color for a node, considering CSS and defaults
func (cr *CanvasRenderer) getEffectiveColor(node *RenderNode) color.Color {
	if node.ComputedStyle != nil && node.ComputedStyle.Color != nil {
		return node.ComputedStyle.Color
	}
	return color.Black
}

// getEffectiveTextStyle returns the text style for a node, considering CSS and defaults
func (cr *CanvasRenderer) getEffectiveTextStyle(node *RenderNode) fyne.TextStyle {
	style := fyne.TextStyle{}
	
	// Check CSS font-weight
	if node.ComputedStyle != nil && node.ComputedStyle.FontWeight == "bold" {
		style.Bold = true
	}
	
	// Check parent tag for style hints
	if node.Parent != nil {
		parentStyle := cr.fontMetrics.GetTextStyle(node.Parent.TagName)
		if parentStyle.Bold {
			style.Bold = true
		}
		if parentStyle.Italic {
			style.Italic = true
		}
		if parentStyle.Monospace {
			style.Monospace = true
		}
	}
	
	return style
}
