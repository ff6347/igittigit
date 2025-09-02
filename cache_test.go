package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheFunctions(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("GetCacheDir", func(t *testing.T) {
		cacheDir, err := getCacheDir()
		if err != nil {
			t.Fatalf("Failed to get cache dir: %v", err)
		}
		
		expected := filepath.Join(tempDir, ".config", "igittigit")
		if cacheDir != expected {
			t.Errorf("Expected cache dir %s, got %s", expected, cacheDir)
		}
		
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			t.Error("Cache directory was not created")
		}
	})

	t.Run("SaveAndLoadCache", func(t *testing.T) {
		templates := []string{"Go", "Node", "Python", "Java"}
		
		err := saveCache(templates)
		if err != nil {
			t.Fatalf("Failed to save cache: %v", err)
		}
		
		cache, err := loadCache()
		if err != nil {
			t.Fatalf("Failed to load cache: %v", err)
		}
		
		if cache == nil {
			t.Fatal("Cache should not be nil")
		}
		
		if len(cache.Templates) != len(templates) {
			t.Errorf("Expected %d templates, got %d", len(templates), len(cache.Templates))
		}
		
		for i, template := range cache.Templates {
			if template != templates[i] {
				t.Errorf("Expected template %s, got %s", templates[i], template)
			}
		}
		
		if cache.LastUpdated == "" {
			t.Error("LastUpdated should not be empty")
		}
	})

	t.Run("IsCacheValid", func(t *testing.T) {
		templates := []string{"Go", "Node"}
		
		err := saveCache(templates)
		if err != nil {
			t.Fatalf("Failed to save cache: %v", err)
		}
		
		cache, err := loadCache()
		if err != nil {
			t.Fatalf("Failed to load cache: %v", err)
		}
		
		if !isCacheValid(cache) {
			t.Error("Fresh cache should be valid")
		}
		
		oldCache := &CacheData{
			Templates:   templates,
			LastUpdated: time.Now().Add(-25 * time.Hour).Format(time.RFC3339),
		}
		
		if isCacheValid(oldCache) {
			t.Error("25-hour old cache should be invalid")
		}
		
		if isCacheValid(nil) {
			t.Error("Nil cache should be invalid")
		}
	})

	t.Run("LoadNonexistentCache", func(t *testing.T) {
		tempDir2 := t.TempDir()
		os.Setenv("HOME", tempDir2)
		
		cache, err := loadCache()
		if err != nil {
			t.Fatalf("Loading nonexistent cache should not error: %v", err)
		}
		
		if cache != nil {
			t.Error("Nonexistent cache should return nil")
		}
	})

	t.Run("InvalidCacheFile", func(t *testing.T) {
		cacheFile, err := getCacheFile()
		if err != nil {
			t.Fatalf("Failed to get cache file path: %v", err)
		}
		
		err = os.WriteFile(cacheFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid cache file: %v", err)
		}
		
		_, err = loadCache()
		if err == nil {
			t.Error("Loading invalid JSON should return error")
		}
	})
}