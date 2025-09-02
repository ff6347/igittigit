package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type CacheData struct {
	Templates   []string `json:"templates"`
	LastUpdated string   `json:"last_updated"`
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	cacheDir := filepath.Join(homeDir, ".config", "igittigit")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", err
	}
	
	return cacheDir, nil
}

func getCacheFile() (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	
	return filepath.Join(cacheDir, "cache.json"), nil
}

func loadCache() (*CacheData, error) {
	cacheFile, err := getCacheFile()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	
	var cache CacheData
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return nil, err
	}
	
	return &cache, nil
}

func saveCache(templates []string) error {
	cacheFile, err := getCacheFile()
	if err != nil {
		return err
	}
	
	cache := CacheData{
		Templates:   templates,
		LastUpdated: time.Now().Format(time.RFC3339),
	}
	
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(cacheFile, data, 0644)
}

func isCacheValid(cache *CacheData) bool {
	if cache == nil {
		return false
	}
	
	lastUpdated, err := time.Parse(time.RFC3339, cache.LastUpdated)
	if err != nil {
		return false
	}
	
	return time.Since(lastUpdated) < 24*time.Hour
}