// ABOUTME: Resolves literal targets and confines local ignore destinations to working trees.
// ABOUTME: Rejects linked-directory traversal and Git administrative paths before writes.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveTarget(cwd, input string) (string, bool, error) {
	if strings.ContainsAny(input, "\r\n\x00") {
		return "", false, fmt.Errorf("path cannot contain line breaks or NUL")
	}
	// Inspect the supplied components before cleaning away any '..' traversal.
	path := cwd
	if filepath.IsAbs(input) {
		path = string(filepath.Separator)
	}
	parts := strings.Split(input, string(filepath.Separator))
	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if strings.EqualFold(part, ".git") {
			return "", false, fmt.Errorf("Git administrative paths are not supported")
		}
		path = filepath.Join(path, part)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", false, err
		}
		if info.Mode()&os.ModeSymlink != 0 && i < len(parts)-1 {
			return "", false, fmt.Errorf("path traverses a symbolic link directory; supply the destination's actual path: %s", path)
		}
	}
	target := filepath.Clean(path)
	if strings.ContainsAny(target, "\r\n\x00") {
		return "", false, fmt.Errorf("resolved path cannot contain line breaks or NUL")
	}
	directory := strings.HasSuffix(input, string(filepath.Separator))
	info, err := os.Lstat(target)
	if err == nil {
		directory = info.IsDir()
	} else if !os.IsNotExist(err) {
		return "", false, err
	}
	return target, directory, nil
}

func existingDirectory(path string) (string, error) {
	for {
		info, err := os.Stat(path)
		if err == nil {
			if !info.IsDir() {
				return "", fmt.Errorf("not a directory: %s", path)
			}
			return path, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		path = parent
	}
}

func targetRepository(target string) (string, error) {
	start := filepath.Dir(target)
	if info, err := os.Lstat(target); err == nil && info.IsDir() {
		start = target
	}
	start, err := existingDirectory(start)
	if err != nil {
		return "", err
	}
	return repositoryRoot(start)
}

func descendant(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func localDestination(cwd, target, root string, here bool) (string, bool, error) {
	if here {
		if !descendant(cwd, target) {
			return "", false, fmt.Errorf("--here requires a descendant of the working directory")
		}
		if cwd == target {
			return "", false, fmt.Errorf("an ignore file cannot ignore its directory itself")
		}
		cwdRoot, err := repositoryRoot(cwd)
		if err != nil {
			return "", false, err
		}
		if cwdRoot != root {
			return "", false, fmt.Errorf("--here cannot cross a repository boundary")
		}
		return filepath.Join(cwd, ".gitignore"), false, nil
	}
	if root == "" {
		return "", false, nil
	}
	if target == root {
		return "", false, fmt.Errorf("an ignore file cannot ignore its directory itself")
	}
	for dir := filepath.Dir(target); ; dir = filepath.Dir(dir) {
		if !descendant(root, dir) {
			return "", false, fmt.Errorf("target crosses repository boundary")
		}
		path := filepath.Join(dir, ".gitignore")
		_, err := os.Lstat(path)
		if err == nil {
			return path, false, nil
		}
		if !os.IsNotExist(err) {
			return "", false, err
		}
		if dir == root {
			return path, true, nil
		}
	}
}
