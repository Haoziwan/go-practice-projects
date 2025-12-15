package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
)

// URLStore manages URL mappings
type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
}

// NewURLStore creates a new URL store
func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]string),
	}
}

// GenerateShortKey generates a random short key
func (s *URLStore) GenerateShortKey() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 and make URL-safe
	key := base64.URLEncoding.EncodeToString(b)
	// Take first 8 characters
	if len(key) > 8 {
		key = key[:8]
	}

	return key, nil
}

// Save stores a URL mapping
func (s *URLStore) Save(shortKey, longURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if shortKey == "" || longURL == "" {
		return errors.New("short key and long URL cannot be empty")
	}

	s.urls[shortKey] = longURL
	return nil
}

// Get retrieves a long URL by short key
func (s *URLStore) Get(shortKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	longURL, exists := s.urls[shortKey]
	if !exists {
		return "", errors.New("short URL not found")
	}

	return longURL, nil
}
