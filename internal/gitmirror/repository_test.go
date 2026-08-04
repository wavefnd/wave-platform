package gitmirror

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitRepositoryReadsBareMirror(t *testing.T) {
	root := t.TempDir()
	working := filepath.Join(root, "working")
	if err := os.MkdirAll(filepath.Join(working, "src"), 0o750); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, root, "init", "-b", "master", working)
	runTestGit(t, working, "config", "user.name", "Wave Test")
	runTestGit(t, working, "config", "user.email", "test@wave.local")
	if err := os.WriteFile(filepath.Join(working, "README.md"), []byte("# Wave\n\nMirror fixture.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(working, "src", "main.wave"), []byte("fun main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, working, "add", ".")
	runTestGit(t, working, "commit", "-m", "initial source")
	runTestGit(t, working, "tag", "v0.1.0")

	mirrors := filepath.Join(root, "mirrors")
	if err := os.MkdirAll(mirrors, 0o750); err != nil {
		t.Fatal(err)
	}
	repository := newGitRepository(mirrors)
	definition := Repository{ID: "wave", RemoteURL: working}
	if err := repository.sync(context.Background(), definition); err != nil {
		t.Fatalf("sync: %v", err)
	}
	branch, err := repository.defaultBranch(context.Background(), "wave")
	if err != nil || branch != "master" {
		t.Fatalf("branch = %q, err = %v", branch, err)
	}
	oid, err := repository.resolve(context.Background(), "wave", "HEAD")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	entries, err := repository.tree(context.Background(), "wave", oid, "")
	if err != nil || len(entries) != 2 {
		t.Fatalf("tree = %#v, err = %v", entries, err)
	}
	blob, err := repository.blob(context.Background(), "wave", oid, "src/main.wave")
	if err != nil || blob.Content != "fun main() {}\n" {
		t.Fatalf("blob = %#v, err = %v", blob, err)
	}
	raw, rawOID, err := repository.rawBlob(context.Background(), "wave", oid, "src/main.wave")
	if err != nil || string(raw) != "fun main() {}\n" || rawOID != blob.OID {
		t.Fatalf("raw blob = %q, oid = %q, err = %v", raw, rawOID, err)
	}
	detail, err := repository.commitDetail(context.Background(), "wave", oid)
	if err != nil {
		t.Fatalf("commit detail: %v", err)
	}
	if detail.Commit.Subject != "initial source" || detail.Body != "" || len(detail.Files) != 2 || !strings.Contains(detail.Patch, "+fun main() {}") {
		t.Fatalf("commit detail = %#v", detail)
	}
	commits, err := repository.commits(context.Background(), "wave", oid, "src/main.wave", 10)
	if err != nil || len(commits) != 1 || commits[0].Subject != "initial source" {
		t.Fatalf("commits = %#v, err = %v", commits, err)
	}
	tags, err := repository.refs(context.Background(), "wave", "refs/tags")
	if err != nil || len(tags) != 1 || tags[0].Name != "v0.1.0" {
		t.Fatalf("tags = %#v, err = %v", tags, err)
	}
}

func runTestGit(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
}
