// ABOUTME: Uses Git discovery and ignore diagnostics rather than matching patterns itself.
// ABOUTME: Reports tracked paths and the source of effective exclusions or exceptions.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func gitOutput(dir, input string, args ...string) ([]byte, string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stdin = strings.NewReader(input)
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	out, err := cmd.Output()
	return out, diagnostics.String(), err
}

func gitFailure(err error, diagnostic string) error {
	return fmt.Errorf("Git failed: %w: %s", err, strings.TrimSpace(diagnostic))
}

func repositoryRoot(dir string) (string, error) {
	out, diagnostic, err := gitOutput(dir, "", "rev-parse", "--show-toplevel")
	if err != nil {
		if strings.HasPrefix(diagnostic, "fatal: not a git repository (or any of the parent directories): .git") ||
			strings.HasPrefix(diagnostic, "fatal: not a git repository (or any parent up to mount point ") {
			return "", nil
		}
		return "", gitFailure(err, diagnostic)
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

type ignoreState struct {
	ignored bool
	source  string
	line    string
	pattern string
}

func (s ignoreState) description() string {
	if s.source == "" {
		return "no matching exclusion (a more specific exception may apply)"
	}
	return fmt.Sprintf("%s:%s:%s", s.source, s.line, s.pattern)
}

func checkIgnore(root, target string, directory bool) (ignoreState, error) {
	var state ignoreState
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return state, err
	}
	path := filepath.ToSlash(rel)
	if directory {
		path += "/"
	}
	out, diagnostic, err := gitOutput(root, path+"\x00", "check-ignore", "--no-index", "-z", "-v", "--stdin")
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return state, gitFailure(err, diagnostic)
		}
	}
	if diagnostic != "" {
		return state, fmt.Errorf("Git ignore diagnostic: %s", strings.TrimSpace(diagnostic))
	}
	if len(out) == 0 {
		return state, nil
	}
	fields := bytes.Split(out, []byte{0})
	if len(fields) != 5 {
		return state, fmt.Errorf("unexpected Git ignore result")
	}
	state.source, state.line, state.pattern = string(fields[0]), string(fields[1]), string(fields[2])
	state.ignored = !strings.HasPrefix(state.pattern, "!")
	return state, nil
}

func trackedPath(root, target string) (bool, error) {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false, err
	}
	out, diagnostic, err := gitOutput(root, "", "--literal-pathspecs", "ls-files", "-z", "--", filepath.ToSlash(rel))
	if err != nil {
		return false, gitFailure(err, diagnostic)
	}
	return len(out) > 0, nil
}
