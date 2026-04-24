package lint_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	lintcmd "github.com/ben-wangz/forgekit/internal/lint"
)

func TestLintRunUsesGitRootLintYAMLByDefault(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".git", "keep"), "")
	writeFile(t, filepath.Join(root, "lint.yaml"), strings.TrimSpace(`
commands:
  - name: "shell check"
    cmd: "sh"
    args: ["-c", "true"]
max_lines_by_ext:
  .go: 5
include:
  - "*.go"
`)+"\n")
	writeFile(t, filepath.Join(root, "ok.go"), "package main\n\nfunc main() {}\n")

	workDir := filepath.Join(root, "pkg")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("create work dir: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := lintcmd.Run(nil, ""); err != nil {
		t.Fatalf("lint run failed: %v", err)
	}
}

func TestLintRunFailsWhenFileExceedsLineLimit(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "lint.yaml"), strings.TrimSpace(`
max_lines_by_ext:
  .go: 2
include:
  - "*.go"
`)+"\n")
	writeFile(t, filepath.Join(root, "too_long.go"), "package main\n\nfunc main() {}\n")

	err := lintcmd.Run(nil, root)
	if err == nil {
		t.Fatal("expected lint run to fail")
	}

	if !strings.Contains(err.Error(), "some files exceed line limits") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
