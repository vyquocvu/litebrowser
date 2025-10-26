package image

// Loader is an interface for loading images.
type Loader interface {
	Load(source string) (*ImageData, error)
	SetOnLoadCallback(callback OnLoadCallback)
	GetCache() *Cache
}
