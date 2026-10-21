// ABOUTME: Exercises informed global-pattern consent using isolated Git configuration.
// ABOUTME: Verifies default destinations, symlink refusals, and cross-repository effects.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobalConsent(t *testing.T) {
	base := isolatedGit(t)
	destination := filepath.Join(base, "xdg", "git", "ignore")
	for _, input := range []string{"", "\n", "no\n", "maybe\n"} {
		out := cli(t, base, input, true, "somewhere/build/")
		contains(t, out, destination)
		contains(t, out, "build/")
		contains(t, out, "every repository")
		contains(t, out, "[y/N]")
		contains(t, out, "Cancelled")
		absent(t, filepath.Join(base, "xdg"))
		absent(t, filepath.Join(base, "somewhere"))
	}
	out := cli(t, base, "yes\n", true, "somewhere/build/")
	contains(t, out, "Rule written")
	if got := contents(t, destination); got != "build/\n" {
		t.Fatal(got)
	}
	for _, name := range []string{"one", "two"} {
		repo := repository(t, base, name)
		if !ignored(t, repo, "nested/build/file") {
			t.Fatal("global rule not applied")
		}
		if ignored(t, repo, "nested/building/file") {
			t.Fatal("global rule too broad")
		}
	}
}

func TestGlobalDestinations(t *testing.T) {
	for _, mode := range []string{"xdg", "home", "configured"} {
		t.Run(mode, func(t *testing.T) {
			base := isolatedGit(t)
			destination := filepath.Join(base, "xdg", "git", "ignore")
			switch mode {
			case "home":
				t.Setenv("XDG_CONFIG_HOME", "")
				destination = filepath.Join(base, ".config", "git", "ignore")
			case "configured":
				destination = filepath.Join(base, "global rules", "ignore")
				gitRun(t, base, "config", "--global", "core.excludesFile", "~/global rules/ignore")
			}
			out := cli(t, base, "y\n", true, "elsewhere/a*b ")
			contains(t, out, destination)
			contains(t, out, "a\\*b\\ ")
			if got := contents(t, destination); got != "a\\*b\\ \n" {
				t.Fatal(got)
			}
			absent(t, filepath.Join(base, "elsewhere"))
		})
	}
}

func TestGlobalAllowPrecedence(t *testing.T) {
	base := isolatedGit(t)
	destination := filepath.Join(base, "xdg", "git", "ignore")
	put(t, destination, "*.env\n")
	repo := repository(t, base, "repo")
	put(t, filepath.Join(repo, ".gitignore"), "/example.env\n")
	out := cli(t, base, "y\n", true, "--allow", "somewhere/example.env")
	contains(t, out, "cannot override repository-local")
	if strings.Contains(out, "state achieved") {
		t.Fatal(out)
	}
	if got := contents(t, destination); got != "*.env\n!example.env\n" {
		t.Fatal(got)
	}
	if !ignored(t, repo, "example.env") {
		t.Fatal("global exception overrode local exclusion")
	}
	other := repository(t, base, "other")
	if ignored(t, other, "example.env") {
		t.Fatal("global exception not applied")
	}
	if !ignored(t, other, "different.env") {
		t.Fatal("global exclusions changed")
	}
	cli(t, base, "y\n", true, "--allow", "folder/")
	contains(t, contents(t, destination), "!folder/\n")
}

func TestGlobalSymlinkAndHere(t *testing.T) {
	base := isolatedGit(t)
	referent := filepath.Join(base, "referent")
	put(t, referent, "unchanged\n")
	destination := filepath.Join(base, "link")
	if err := os.Symlink(referent, destination); err != nil {
		t.Fatal(err)
	}
	gitRun(t, base, "config", "--global", "core.excludesFile", destination)
	contains(t, cli(t, base, "y\n", false, "new"), "symbolic link")
	if got := contents(t, referent); got != "unchanged\n" {
		t.Fatal(got)
	}
	out := cli(t, base, "", true, "--here", "new")
	if strings.Contains(out, "every repository") {
		t.Fatal(out)
	}
	if got := contents(t, filepath.Join(base, ".gitignore")); got != "/new\n" {
		t.Fatal(got)
	}
	gitRun(t, base, "config", "--global", "core.excludesFile", "~/missing/ignore")
	cli(t, base, "no\n", true, "new")
	absent(t, filepath.Join(base, "missing"))
}

func TestFilesystemBoundaryDiscovery(t *testing.T) {
	base := isolatedGit(t)
	t.Setenv("GIT_DISCOVERY_ACROSS_FILESYSTEM", "0")
	out, err := exec.Command("git", "-C", "/dev", "rev-parse", "--show-toplevel").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "Stopping at filesystem boundary") {
		t.Skip("no /dev filesystem boundary available")
	}
	contains(t, cli(t, base, "no\n", true, "/dev/future"), "every repository")
	absent(t, filepath.Join(base, "xdg"))
}

func TestGitDiscoveryErrors(t *testing.T) {
	base := isolatedGit(t)
	put(t, filepath.Join(base, "gitconfig"), "invalid config [\n")
	contains(t, cli(t, base, "y\n", false, "new"), "Git failed")
	absent(t, filepath.Join(base, "xdg"))
	absent(t, filepath.Join(base, ".gitignore"))
}
