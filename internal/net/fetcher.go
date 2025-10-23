package net

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Fetcher handles HTTP requests
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new Fetcher instance
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{},
	}
}

// Fetch retrieves the content from the given URL
func (f *Fetcher) Fetch(url string) (string, error) {
	return f.FetchWithContext(context.Background(), url)
}

// FetchWithContext retrieves the content from the given URL with cancellation support
func (f *Fetcher) FetchWithContext(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
