package image

import (
	"container/list"
	"sync"
)

// Cache implements an LRU (Least Recently Used) cache for images
type Cache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	lruList  *list.List
}

// cacheEntry represents an entry in the cache
type cacheEntry struct {
	key   string
	value *ImageData
}

// NewCache creates a new LRU cache with the specified capacity
func NewCache(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	return &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get retrieves an image from the cache
func (c *Cache) Get(key string) *ImageData {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// Move to front (most recently used)
		c.lruList.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value
	}
	return nil
}

// Put adds or updates an image in the cache
func (c *Cache) Put(key string, value *ImageData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if elem, ok := c.items[key]; ok {
		// Update existing entry and move to front
		c.lruList.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	// Add new entry
	entry := &cacheEntry{key: key, value: value}
	elem := c.lruList.PushFront(entry)
	c.items[key] = elem

	// Evict least recently used if over capacity
	if c.lruList.Len() > c.capacity {
		c.evict()
	}
}

// evict removes the least recently used item from the cache
func (c *Cache) evict() {
	elem := c.lruList.Back()
	if elem != nil {
		c.lruList.Remove(elem)
		entry := elem.Value.(*cacheEntry)
		delete(c.items, entry.key)
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList = list.New()
}

// Len returns the current number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// SetCapacity updates the cache capacity
func (c *Cache) SetCapacity(capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if capacity <= 0 {
		capacity = 100
	}
	c.capacity = capacity

	// Evict items if necessary
	for c.lruList.Len() > c.capacity {
		c.evict()
	}
}
