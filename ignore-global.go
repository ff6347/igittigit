// ABOUTME: Offers basename-based global ignore patterns only with informed consent.
// ABOUTME: Resolves Git's excludes file without changing configuration or assuming a filename.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func globalDestination(cwd string) (string, error) {
	value, diagnostic, err := gitOutput(cwd, "", "config", "--path", "--get", "core.excludesFile")
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return "", gitFailure(err, diagnostic)
		}
	}
	if len(value) > 0 {
		path := strings.TrimSuffix(string(value), "\n")
		if path == "" {
			return "", fmt.Errorf("core.excludesFile is empty; configure an ignore file before using global patterns")
		}
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("core.excludesFile is relative and resolves differently in each repository; configure an absolute path for global patterns")
		}
		return filepath.Clean(path), nil
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdg = filepath.Join(home, ".config")
	}
	if !filepath.IsAbs(xdg) {
		return "", fmt.Errorf("global patterns require an absolute XDG_CONFIG_HOME or HOME directory")
	}
	return filepath.Join(xdg, "git", "ignore"), nil
}

func offerGlobal(cwd, target string, directory, allow bool, in io.Reader, out io.Writer) error {
	if filepath.Dir(target) == target {
		return fmt.Errorf("filesystem root has no literal basename for a global pattern")
	}
	destination, err := globalDestination(cwd)
	if err != nil {
		return err
	}
	if err := safeIgnoreFile(destination); err != nil {
		return err
	}
	rule := literalRule(filepath.Base(target), directory, allow)
	fmt.Fprintln(out, "No containing working tree. A global pattern affects every repository, not arbitrary filesystem locations.")
	fmt.Fprintln(out, "This basename pattern matches at any depth across all repositories.")
	fmt.Fprintln(out, "Global exceptions cannot override repository-local exclusions or ignored parents.")
	if !confirm(in, out, fmt.Sprintf("Append exact global pattern %s to %s?", rule, destination)) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create global ignore directory: %w", err)
	}
	return applyRule(destination, rule, target, "", directory, allow, out)
}
