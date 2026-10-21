// ABOUTME: Checks configuration cases that cannot supply a repository-independent ignore file.
// ABOUTME: Verifies refusal leaves both configuration and candidate destinations unchanged.
package main

import (
	"path/filepath"
	"testing"
)

func TestRelativeXDGDestination(t *testing.T) {
	base := isolatedGit(t)
	t.Setenv("XDG_CONFIG_HOME", "relative-xdg")
	contains(t, cli(t, base, "yes\n", false, "build/"), "absolute")
	absent(t, filepath.Join(base, "relative-xdg"))
}

func TestRelativeGlobalDestination(t *testing.T) {
	base := isolatedGit(t)
	gitRun(t, base, "config", "--global", "core.excludesFile", "relative ignore")
	before := contents(t, filepath.Join(base, "gitconfig"))
	contains(t, cli(t, base, "yes\n", false, "build/"), "absolute")
	absent(t, filepath.Join(base, "relative ignore"))
	if got := contents(t, filepath.Join(base, "gitconfig")); got != before {
		t.Fatal("configuration changed")
	}
}
