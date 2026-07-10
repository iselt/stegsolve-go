package imagecore

import (
	"fmt"
	"sync"
)

// Store holds at most one loaded image for the app session.
type Store struct {
	mu    sync.RWMutex
	image *LoadedImage
}

// NewStore creates an empty image store.
func NewStore() *Store {
	return &Store{}
}

// Set replaces the current image.
func (s *Store) Set(img *LoadedImage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.image = img
}

// Get returns the current image or an error if none / id mismatch.
func (s *Store) Get(imageID string) (*LoadedImage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.image == nil {
		return nil, fmt.Errorf("尚未加载图像")
	}
	if imageID != "" && s.image.ID != imageID {
		return nil, fmt.Errorf("imageId 无效或已失效")
	}
	return s.image, nil
}

// Current returns the loaded image without id check (may be nil).
func (s *Store) Current() *LoadedImage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.image
}
