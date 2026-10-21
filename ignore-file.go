// ABOUTME: Appends ignore rules while preserving bytes, permissions, and line endings.
// ABOUTME: Validates file identity before writing and exclusively creates missing files.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func safeIgnoreFile(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing symbolic link ignore file: %s", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("ignore destination is not a regular file: %s", path)
	}
	return nil
}

func appendRule(path string, data []byte, rule string) error {
	if err := safeIgnoreFile(path); err != nil {
		return err
	}
	ending := "\n"
	if strings.Contains(string(data), "\r\n") {
		ending = "\r\n"
	}
	addition := rule + ending
	if len(data) > 0 && data[len(data)-1] != '\n' {
		addition = ending + addition
	}
	expected, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	file, err := openIgnoreFile(path, expected)
	if err != nil {
		return fmt.Errorf("open ignore file: %w", err)
	}
	_, writeErr := io.WriteString(file, addition)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write ignore file: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close ignore file: %w", closeErr)
	}
	return nil
}

func openIgnoreFile(path string, expected os.FileInfo) (*os.File, error) {
	flags := os.O_WRONLY | os.O_APPEND
	if expected == nil {
		flags |= os.O_CREATE | os.O_EXCL
	}
	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err == nil && expected != nil && !os.SameFile(expected, opened) {
		err = fmt.Errorf("ignore file changed during validation: %s", path)
	}
	if err == nil {
		current, statErr := os.Lstat(path)
		switch {
		case statErr != nil:
			err = statErr
		case !current.Mode().IsRegular() || !os.SameFile(current, opened):
			err = fmt.Errorf("ignore file replaced or is a symbolic link: %s", path)
		}
	}
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return file, nil
}
