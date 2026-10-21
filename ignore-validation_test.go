// ABOUTME: Covers path validation and preservation of filename bytes before writes.
// ABOUTME: Prevents inherited path components from injecting ignore-file lines.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInheritedLineBreaks(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	path := filepath.Join(repo, ".gitignore")
	put(t, path, "# keep\n")
	cwd := filepath.Join(repo, "bad\nname")
	if err := os.Mkdir(cwd, 0755); err != nil {
		t.Fatal(err)
	}
	contains(t, cli(t, cwd, "", false, "secret"), "line breaks")
	if got := contents(t, path); got != "# keep\n" {
		t.Fatal(got)
	}
	outside := filepath.Join(base, "bad\rname")
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	contains(t, cli(t, outside, "yes\n", false, "."), "line breaks")
	absent(t, filepath.Join(base, "xdg"))
}

func TestLiteralFilenameBytes(t *testing.T) {
	name := "/bad\xff*"
	if got := literalRule(name, false, false); got != "/bad\xff\\*" {
		t.Fatalf("rule bytes not preserved: %q", got)
	}
}
