// ABOUTME: Verifies ignore-file opens cannot redirect writes to a replacement symlink.
// ABOUTME: Exercises real filesystem replacements between validation and opening.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreFileReplacement(t *testing.T) {
	base := isolatedGit(t)
	path := filepath.Join(base, ".gitignore")
	referent := filepath.Join(base, "referent")
	put(t, path, "original\n")
	put(t, referent, "untouched\n")
	expected, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(referent, path); err != nil {
		t.Fatal(err)
	}
	file, err := openIgnoreFile(path, expected)
	if err == nil {
		file.Close()
		t.Fatal("replacement link was accepted")
	}
	if got := contents(t, referent); got != "untouched\n" {
		t.Fatal(got)
	}
	file, err = openIgnoreFile(path, nil)
	if err == nil {
		file.Close()
		t.Fatal("creation followed existing link")
	}
	if got := contents(t, referent); got != "untouched\n" {
		t.Fatal(got)
	}
}
