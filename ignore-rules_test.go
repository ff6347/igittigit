// ABOUTME: Verifies literal pattern semantics and byte-preserving ignore writes with Git.
// ABOUTME: Exercises effective exceptions, tracked paths, and overriding ignore sources.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func ignored(t *testing.T, repo, path string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "check-ignore", "--no-index", "-q", "--", path)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return true
	}
	if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 && len(out) == 0 {
		return false
	}
	t.Fatalf("check-ignore %q: %v: %s", path, err, out)
	return false
}

func TestLiteralRules(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	put(t, filepath.Join(repo, ".gitignore"), "")
	for _, name := range []string{"plain", "#hash", "!bang", "a*b", "a?b", "a[b]", "a\\b", " leading", "trailing ", "two words", "é.txt", "tab\tname"} {
		t.Run(name, func(t *testing.T) {
			put(t, filepath.Join(repo, name), "")
			cli(t, repo, "", true, "./"+name)
			if !ignored(t, repo, name) {
				t.Fatal("literal target not ignored")
			}
			for _, other := range []string{"nested/" + name, name + "sibling", "axb"} {
				if ignored(t, repo, other) {
					t.Fatalf("unrelated name %q ignored", other)
				}
			}
		})
	}
	put(t, filepath.Join(repo, "folder", "file"), "")
	cli(t, repo, "", true, "folder")
	contains(t, contents(t, filepath.Join(repo, ".gitignore")), "/folder/\n")
	before := contents(t, filepath.Join(repo, ".gitignore"))
	for _, path := range []string{"bad\nname", "bad\rname"} {
		contains(t, cli(t, repo, "", false, path), "line breaks")
	}
	if after := contents(t, filepath.Join(repo, ".gitignore")); after != before {
		t.Fatal("invalid input changed ignore file")
	}
}

func TestPreserveAndRepeat(t *testing.T) {
	for _, initial := range []string{"", "# comment", "# comment\n", "# comment\r\n", "# comment\r\nlast"} {
		t.Run(initial, func(t *testing.T) {
			base := isolatedGit(t)
			repo := repository(t, base, "repo")
			path := filepath.Join(repo, ".gitignore")
			put(t, path, initial)
			if err := os.Chmod(path, 0600); err != nil {
				t.Fatal(err)
			}
			cli(t, repo, "", true, "new")
			ending := "\n"
			if strings.Contains(initial, "\r\n") {
				ending = "\r\n"
			}
			want := initial
			if initial != "" && !strings.HasSuffix(initial, "\n") {
				want += ending
			}
			want += "/new" + ending
			if got := contents(t, path); got != want {
				t.Fatalf("got %q want %q", got, want)
			}
			contains(t, cli(t, repo, "", true, "new"), "No write needed")
			if got := contents(t, path); got != want {
				t.Fatal(got)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("mode: %v %v", info, err)
			}
		})
	}
}

func TestEffectiveRules(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	path := filepath.Join(repo, ".gitignore")
	put(t, path, "/name\n!/name\n")
	cli(t, repo, "", true, "name")
	if got := contents(t, path); got != "/name\n!/name\n/name\n" {
		t.Fatal(got)
	}
	cli(t, repo, "", true, "--allow", "name")
	if ignored(t, repo, "name") {
		t.Fatal("exception ineffective")
	}
	before := contents(t, path)
	contains(t, cli(t, repo, "", true, "--allow", "name"), "No write needed")
	if got := contents(t, path); got != before {
		t.Fatal(got)
	}
	put(t, path, "/config/\n")
	out := cli(t, repo, "", false, "--allow", "config/example.env")
	contains(t, out, "Rule written")
	contains(t, out, "ineffective")
	contains(t, out, ".gitignore:1:/config/")
	if got := contents(t, path); got != "/config/\n!/config/example.env\n" {
		t.Fatal(got)
	}
	if !ignored(t, repo, "config/example.env") {
		t.Fatal("parent exclusion rewritten")
	}
	put(t, path, "")
	put(t, filepath.Join(repo, "app", ".gitignore"), "!/cache\n")
	contains(t, cli(t, repo, "", false, "--here", "app/cache"), "blocking rule: app/.gitignore:1:!/cache")
	put(t, path, "/config/\n")
	put(t, filepath.Join(repo, "config", ".gitignore"), "/secret\n")
	cli(t, repo, "", true, "--allow", "config/")
	if !ignored(t, repo, "config/secret") {
		t.Fatal("directory exception cleared descendant exclusion")
	}
}

func TestTrackedAndOtherSources(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	put(t, filepath.Join(repo, ".gitignore"), "")
	put(t, filepath.Join(repo, "tracked"), "data")
	gitRun(t, repo, "add", "tracked")
	before := gitRun(t, repo, "ls-files", "--stage")
	contains(t, cli(t, repo, "", true, "tracked"), "do not untrack")
	if after := gitRun(t, repo, "ls-files", "--stage"); after != before {
		t.Fatal("index changed")
	}
	// Repeated commands must still warn about tracking, even without a write.
	contains(t, cli(t, repo, "", true, "tracked"), "do not untrack")
	put(t, filepath.Join(repo, ".git", "info", "exclude"), "/excluded\n")
	cli(t, repo, "", true, "--allow", "excluded")
	if ignored(t, repo, "excluded") {
		t.Fatal("local exception did not override info/exclude")
	}
}

func TestWriteFailures(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	path := filepath.Join(repo, ".gitignore")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	contains(t, cli(t, repo, "", false, "x"), "not a regular file")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	put(t, path, "keep\n")
	if err := os.Chmod(path, 0444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(path, 0644) })
	contains(t, cli(t, repo, "", false, "x"), "permission denied")
	if got := contents(t, path); got != "keep\n" {
		t.Fatal(got)
	}
}
