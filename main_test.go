package main

import (
	"os"
	"testing"
)

func TestHasFzf(t *testing.T) {
	result := hasFzf()
	t.Logf("fzf available: %v", result)
}

func TestFileExists(t *testing.T) {
	tempFile := t.TempDir() + "/test.txt"
	
	if fileExists(tempFile) {
		t.Error("Non-existent file should not exist")
	}
	
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if !fileExists(tempFile) {
		t.Error("Existing file should exist")
	}
}

func TestPromptForActionValidation(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"o", "overwrite"},
		{"O", "overwrite"},
		{"overwrite", "overwrite"},
		{"a", "append"},
		{"A", "append"},
		{"append", "append"},
		{"c", "cancel"},
		{"C", "cancel"},
		{"cancel", "cancel"},
		{"", "cancel"},
		{"invalid", "cancel"},
		{"x", "cancel"},
	}
	
	for _, tc := range testCases {
		t.Run("input_"+tc.input, func(t *testing.T) {
			t.Logf("Would test input '%s' expecting '%s'", tc.input, tc.expected)
		})
	}
}

func TestPrintUsage(t *testing.T) {
	t.Log("Testing printUsage function exists and can be called")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printUsage() panicked: %v", r)
		}
	}()
	
	printUsage()
}