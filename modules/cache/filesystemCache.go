package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheItem struct {
    MimeType   string
    Expiration time.Time
}

type FilesystemCache struct {
    cacheDir string
    items    map[string]CacheItem
	mutex    sync.RWMutex
}

func NewFilesystemCache(cacheDir string) *FilesystemCache {
    return &FilesystemCache{
        cacheDir: cacheDir,
        items:    make(map[string]CacheItem),
		mutex:    sync.RWMutex{},
    }
}

func (fc *FilesystemCache) Get(cacheKey string) ([]byte, string, error) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()

	item, ok := fc.items[cacheKey]

	if !ok || time.Now().After(item.Expiration) {
		return nil, "", fmt.Errorf("cache item not found or expired")
	}

	cachedImagePath := filepath.Join(fc.cacheDir, cacheKey)
	imageBytes, err := os.ReadFile(cachedImagePath)
	if err != nil {
		return nil, "", err
	}

	mimeType := item.MimeType
	return imageBytes, mimeType, nil
}

func (fc *FilesystemCache) Set(cacheKey string, value []byte, mimeType string, expiration time.Duration) error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	cachedImagePath := filepath.Join(fc.cacheDir, cacheKey)

	if err := os.MkdirAll(filepath.Dir(cachedImagePath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(cachedImagePath, value, 0644); err != nil {
		return err
	}

	fc.items[cacheKey] = CacheItem{
		MimeType:   mimeType,
		Expiration: time.Now().Add(expiration),
	}

	return nil
}