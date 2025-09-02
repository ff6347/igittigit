package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type TemplateResponse struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

func hasGhCli() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func executeGhCommand(endpoint string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint, "-H", "X-GitHub-Api-Version: 2022-11-28")
	return cmd.Output()
}

func makeHttpRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	return io.ReadAll(resp.Body)
}

func fetchFromGitHub(endpoint string) ([]byte, error) {
	if hasGhCli() {
		data, err := executeGhCommand(endpoint)
		if err == nil {
			return data, nil
		}
	}
	
	url := "https://api.github.com" + endpoint
	return makeHttpRequest(url)
}

func getTemplateList() ([]string, error) {
	cache, err := loadCache()
	if err == nil && isCacheValid(cache) {
		return cache.Templates, nil
	}
	
	data, err := fetchFromGitHub("/gitignore/templates")
	if err != nil {
		if cache != nil {
			return cache.Templates, nil
		}
		return nil, fmt.Errorf("failed to fetch templates: %w", err)
	}
	
	var templates []string
	err = json.Unmarshal(data, &templates)
	if err != nil {
		if cache != nil {
			return cache.Templates, nil
		}
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	
	saveCache(templates)
	return templates, nil
}

func getTemplate(name string) (string, error) {
	endpoint := fmt.Sprintf("/gitignore/templates/%s", name)
	data, err := fetchFromGitHub(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to fetch template '%s': %w", name, err)
	}
	
	var response TemplateResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse template '%s': %w", name, err)
	}
	
	return strings.TrimSpace(response.Source), nil
}