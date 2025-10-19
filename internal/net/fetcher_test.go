package net

import (
	"testing"
)

func TestFetcher(t *testing.T) {
	fetcher := NewFetcher()
	if fetcher == nil {
		t.Fatal("NewFetcher() returned nil")
	}
	if fetcher.client == nil {
		t.Fatal("Fetcher client is nil")
	}
}

func TestFetchInvalidURL(t *testing.T) {
	fetcher := NewFetcher()
	_, err := fetcher.Fetch("not-a-url")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
