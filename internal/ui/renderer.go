package ui

import "fyne.io/fyne/v2"

type HTMLRenderer interface {
	RenderHTML(htmlContent string) (fyne.CanvasObject, error)
	SetCurrentURL(url string)
	ResolveURL(url string) string
	SetWindow(w fyne.Window)
	SetNavigationCallback(callback func(url string))
}