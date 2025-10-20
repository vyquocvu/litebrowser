package ui

import "sync"

// BrowserState manages the browser's navigation history and bookmarks
type BrowserState struct {
	mu             sync.RWMutex
	history        []string
	currentIndex   int
	bookmarks      []string
}

// NewBrowserState creates a new browser state
func NewBrowserState() *BrowserState {
	return &BrowserState{
		history:      make([]string, 0),
		currentIndex: -1,
		bookmarks:    make([]string, 0),
	}
}

// AddToHistory adds a URL to navigation history
func (s *BrowserState) AddToHistory(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Remove forward history if we're navigating to a new page from the middle
	if s.currentIndex < len(s.history)-1 {
		s.history = s.history[:s.currentIndex+1]
	}
	
	s.history = append(s.history, url)
	s.currentIndex = len(s.history) - 1
}

// CanGoBack returns true if there's a previous page to go back to
func (s *BrowserState) CanGoBack() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentIndex > 0
}

// CanGoForward returns true if there's a next page to go forward to
func (s *BrowserState) CanGoForward() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentIndex < len(s.history)-1
}

// GoBack moves back in history and returns the previous URL
func (s *BrowserState) GoBack() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.currentIndex <= 0 {
		return "", false
	}
	
	s.currentIndex--
	return s.history[s.currentIndex], true
}

// GoForward moves forward in history and returns the next URL
func (s *BrowserState) GoForward() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.currentIndex >= len(s.history)-1 {
		return "", false
	}
	
	s.currentIndex++
	return s.history[s.currentIndex], true
}

// GetCurrentURL returns the current URL
func (s *BrowserState) GetCurrentURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.currentIndex < 0 || s.currentIndex >= len(s.history) {
		return ""
	}
	return s.history[s.currentIndex]
}

// AddBookmark adds a URL to bookmarks
func (s *BrowserState) AddBookmark(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if bookmark already exists
	for _, bookmark := range s.bookmarks {
		if bookmark == url {
			return
		}
	}
	
	s.bookmarks = append(s.bookmarks, url)
}

// RemoveBookmark removes a URL from bookmarks
func (s *BrowserState) RemoveBookmark(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i, bookmark := range s.bookmarks {
		if bookmark == url {
			s.bookmarks = append(s.bookmarks[:i], s.bookmarks[i+1:]...)
			return
		}
	}
}

// GetBookmarks returns a copy of all bookmarks
func (s *BrowserState) GetBookmarks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	bookmarks := make([]string, len(s.bookmarks))
	copy(bookmarks, s.bookmarks)
	return bookmarks
}

// IsBookmarked checks if a URL is bookmarked
func (s *BrowserState) IsBookmarked(url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, bookmark := range s.bookmarks {
		if bookmark == url {
			return true
		}
	}
	return false
}

// GetHistory returns a copy of navigation history
func (s *BrowserState) GetHistory() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	history := make([]string, len(s.history))
	copy(history, s.history)
	return history
}
