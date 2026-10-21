// ABOUTME: Exercises path commands through the compiled CLI in real repositories.
// ABOUTME: Isolates Git configuration and verifies destinations and cancellation.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var cliBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "igittigit-tests-")
	if err != nil {
		panic(err)
	}
	cliBinary = filepath.Join(dir, "igittigit")
	cmd := exec.Command("go", "build", "-o", cliBinary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out) + err.Error())
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func isolatedGit(t *testing.T) string {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_CONFIG", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", base)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(base, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CEILING_DIRECTORIES", base)
	return base
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func repository(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "init", "-q")
	return dir
}

func put(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func contents(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func absent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s absent: %v", path, err)
	}
}

func cli(t *testing.T, cwd, input string, success bool, args ...string) string {
	t.Helper()
	cmd := exec.Command(cliBinary, args...)
	cmd.Dir = cwd
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if (err == nil) != success {
		t.Fatalf("CLI %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func contains(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("output %q does not contain %q", text, want)
	}
}

func TestPathDispatch(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	put(t, filepath.Join(repo, ".gitignore"), "# keep\n")
	for _, args := range [][]string{{"--bad", "x"}, {"--here"}, {"--allow"}, {"a", "b"}, {"x", "--here"}, {"--"}} {
		contains(t, cli(t, repo, "", false, args...), "Error:")
		if got := contents(t, filepath.Join(repo, ".gitignore")); got != "# keep\n" {
			t.Fatal(got)
		}
	}
	cli(t, repo, "", true, "--here", "--allow", "./future")
	contains(t, contents(t, filepath.Join(repo, ".gitignore")), "!/future\n")
	cli(t, repo, "", true, "--", "-cache")
	cli(t, repo, "", true, "./list")
	contains(t, contents(t, filepath.Join(repo, ".gitignore")), "/-cache\n/list\n")
	contains(t, cli(t, repo, "", true, "help"), "Usage:")
	contains(t, cli(t, repo, "", false, "get"), "template name required")
}

func TestPathDestinations(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	rootIgnore := filepath.Join(repo, ".gitignore")
	put(t, rootIgnore, "# root\n")
	nested := filepath.Join(repo, "app")
	put(t, filepath.Join(nested, ".gitignore"), "# app\n")
	cli(t, base, "", true, filepath.Join(nested, "missing", "build")+"/")
	if got := contents(t, filepath.Join(nested, ".gitignore")); got != "# app\n/missing/build/\n" {
		t.Fatal(got)
	}
	absent(t, filepath.Join(nested, "missing"))
	cli(t, repo, "", true, "app/")
	contains(t, contents(t, rootIgnore), "/app/\n")
	cli(t, repo, "", true, "--here", "app/cache")
	contains(t, contents(t, rootIgnore), "/app/cache\n")
	inner := repository(t, repo, "inner")
	contains(t, cli(t, repo, "", false, "--here", "inner/file"), "repository boundary")
	cli(t, repo, "yes\n", true, "inner/new")
	contains(t, contents(t, filepath.Join(inner, ".gitignore")), "/new\n")
	contains(t, cli(t, nested, "", false, "--here", "../outside"), "descendant")
	contains(t, cli(t, repo, "", false, "--here", "."), "itself")
	contains(t, cli(t, repo, "", false, ".git/config"), "administrative")
}

func TestCreationConsent(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	for _, input := range []string{"", "\n", "n\n", "maybe\n"} {
		contains(t, cli(t, repo, input, true, "missing/child"), "Cancelled")
		absent(t, filepath.Join(repo, ".gitignore"))
		absent(t, filepath.Join(repo, "missing"))
	}
	cli(t, repo, "y\n", true, "missing/child")
	if got := contents(t, filepath.Join(repo, ".gitignore")); got != "/missing/child\n" {
		t.Fatal(got)
	}
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	contains(t, cli(t, outside, "", true, "--here", "future/"), "takes effect")
	if got := contents(t, filepath.Join(outside, ".gitignore")); got != "/future/\n" {
		t.Fatal(got)
	}
}

func TestPathSymlinks(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	put(t, filepath.Join(repo, ".gitignore"), "")
	referent := filepath.Join(base, "actual")
	put(t, filepath.Join(referent, "file"), "keep")
	if err := os.Symlink(referent, filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	cli(t, repo, "", true, "link")
	if got := contents(t, filepath.Join(repo, ".gitignore")); got != "/link\n" {
		t.Fatal(got)
	}
	contains(t, cli(t, repo, "", false, "link/file"), "actual path")
	contains(t, cli(t, repo, "", false, "link/../file"), "actual path")
	contains(t, cli(t, repo, "", false, "link/"), "symbolic link")
	if err := os.Remove(filepath.Join(repo, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(referent, "file"), filepath.Join(repo, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	contains(t, cli(t, repo, "", false, "new"), "symbolic link")
	if got := contents(t, filepath.Join(referent, "file")); got != "keep" {
		t.Fatal(got)
	}
}

func TestWorktreeAndGitRequirement(t *testing.T) {
	base := isolatedGit(t)
	repo := repository(t, base, "repo")
	gitRun(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--allow-empty", "-qm", "initial")
	worktree := filepath.Join(base, "worktree")
	gitRun(t, repo, "worktree", "add", "-qb", "other", worktree)
	cli(t, base, "yes\n", true, filepath.Join(worktree, "future"))
	contains(t, contents(t, filepath.Join(worktree, ".gitignore")), "/future\n")
	absent(t, filepath.Join(repo, ".gitignore"))
	t.Setenv("PATH", t.TempDir())
	contains(t, cli(t, repo, "", false, "--here", "x"), "install Git")
	contains(t, cli(t, repo, "", true, "help"), "Usage:")
}
