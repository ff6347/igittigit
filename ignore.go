// ABOUTME: Runs literal path ignore commands with explicit destination selection.
// ABOUTME: Preserves ignore files and reports the effective state determined by Git.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type pathOptions struct {
	path  string
	here  bool
	allow bool
}

func parsePath(args []string) (pathOptions, error) {
	var opts pathOptions
	flags := true
	for _, arg := range args {
		if arg == "" {
			return opts, fmt.Errorf("one nonempty literal path required")
		}
		if flags && strings.HasPrefix(arg, "-") {
			switch arg {
			case "--":
				flags = false
			case "--here":
				opts.here = true
			case "--allow":
				opts.allow = true
			default:
				return opts, fmt.Errorf("unknown flag %q; flags must precede the path", arg)
			}
			continue
		}
		if opts.path != "" {
			return opts, fmt.Errorf("supply exactly one path; flags must precede the path")
		}
		opts.path = arg
		flags = false
	}
	if opts.path == "" {
		return opts, fmt.Errorf("one literal path required")
	}
	return opts, nil
}

func runPath(args []string, in io.Reader, out io.Writer) error {
	opts, err := parsePath(args)
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("path mode requires Git; install Git and ensure it is on PATH: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target, directory, err := resolveTarget(cwd, opts.path)
	if err != nil {
		return err
	}
	root, err := targetRepository(target)
	if err != nil {
		return err
	}
	destination, create, err := localDestination(cwd, target, root, opts.here)
	if err != nil {
		return err
	}
	if root == "" && !opts.here {
		return offerGlobal(cwd, target, directory, opts.allow, in, out)
	}
	if err := safeIgnoreFile(destination); err != nil {
		return err
	}
	rel, err := filepath.Rel(filepath.Dir(destination), target)
	if err != nil {
		return err
	}
	rule := literalRule("/"+filepath.ToSlash(rel), directory, opts.allow)
	if create && !confirm(in, out, fmt.Sprintf("Create %s with rule %s?", destination, rule)) {
		return nil
	}
	if root == "" {
		fmt.Fprintln(out, "This .gitignore takes effect once this directory belongs to a Git working tree.")
	}
	return applyRule(destination, rule, target, root, directory, opts.allow, out)
}

func confirm(in io.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [y/N]: ", question)
	scanner := bufio.NewScanner(in)
	if scanner.Scan() {
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "y", "yes":
			return true
		}
	}
	fmt.Fprintln(out, "Cancelled; no files changed.")
	return false
}

func applyRule(destination, rule, target, root string, directory, allow bool, out io.Writer) error {
	data, err := os.ReadFile(destination)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if root != "" {
		state, err := checkIgnore(root, target, directory)
		if err != nil {
			return err
		}
		tracked, err := trackedPath(root, target)
		if err != nil {
			return err
		}
		if tracked {
			fmt.Fprintln(out, "Warning: target contains tracked files; ignore rules do not untrack them. The index is unchanged.")
		}
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			if line == rule && state.ignored != allow {
				fmt.Fprintf(out, "No write needed: %s already effective in %s\n", rule, destination)
				return nil
			}
		}
	}
	if err := appendRule(destination, data, rule); err != nil {
		return err
	}
	fmt.Fprintf(out, "Rule written to %s: %s\n", destination, rule)
	if root == "" {
		return nil
	}
	state, err := checkIgnore(root, target, directory)
	if err != nil {
		return err
	}
	if state.ignored == allow {
		return fmt.Errorf("rule written, but requested state is ineffective; blocking rule: %s", state.description())
	}
	fmt.Fprintln(out, "Requested ignore state achieved (directory descendants may have their own rules).")
	return nil
}

func literalRule(path string, directory, allow bool) string {
	var b strings.Builder
	if allow {
		b.WriteByte('!')
	}
	for i := 0; i < len(path); i++ {
		c := path[i]
		if strings.ContainsRune("\\*?[]#! ", rune(c)) {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	if directory {
		b.WriteByte('/')
	}
	return b.String()
}
