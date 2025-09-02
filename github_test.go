package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHasGhCli(t *testing.T) {
	result := hasGhCli()
	t.Logf("gh CLI available: %v", result)
}

func TestFetchFromGitHubMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/gitignore/templates" {
			templates := []string{"Go", "Node", "Python"}
			json.NewEncoder(w).Encode(templates)
		} else if r.URL.Path == "/gitignore/templates/Go" {
			response := TemplateResponse{
				Name:   "Go",
				Source: "# Go\n*.exe\n*.so\n*.dylib",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	originalURL := "https://api.github.com"
	
	t.Run("GetTemplateList", func(t *testing.T) {
		data, err := makeHttpRequest(server.URL + "/gitignore/templates")
		if err != nil {
			t.Fatalf("Failed to fetch templates: %v", err)
		}
		
		var templates []string
		err = json.Unmarshal(data, &templates)
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}
		
		expected := []string{"Go", "Node", "Python"}
		if len(templates) != len(expected) {
			t.Errorf("Expected %d templates, got %d", len(expected), len(templates))
		}
		
		for i, template := range templates {
			if template != expected[i] {
				t.Errorf("Expected template %s, got %s", expected[i], template)
			}
		}
	})
	
	t.Run("GetTemplate", func(t *testing.T) {
		data, err := makeHttpRequest(server.URL + "/gitignore/templates/Go")
		if err != nil {
			t.Fatalf("Failed to fetch Go template: %v", err)
		}
		
		var response TemplateResponse
		err = json.Unmarshal(data, &response)
		if err != nil {
			t.Fatalf("Failed to parse Go template: %v", err)
		}
		
		if response.Name != "Go" {
			t.Errorf("Expected template name 'Go', got '%s'", response.Name)
		}
		
		if response.Source == "" {
			t.Error("Expected non-empty source content")
		}
		
		t.Logf("Template content preview: %s", response.Source[:20])
	})

	_ = originalURL
}