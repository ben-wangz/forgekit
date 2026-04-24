package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestForgekitVersionGetBinary(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "version-control.yaml"), "binaries:\n  - name: forgekit\n    path: .\n    versionFile: VERSION\n")
	writeFile(t, filepath.Join(projectRoot, "VERSION"), "1.2.3\n")

	result := runForgekit(t, "--project-root", projectRoot, "version", "get", "forgekit")
	if result.err != nil {
		t.Fatalf("forgekit version get failed: %v\noutput:\n%s", result.err, result.output)
	}

	if strings.TrimSpace(result.output) != "1.2.3" {
		t.Fatalf("unexpected output: %q", result.output)
	}
}

func TestForgekitLintPassesWithExplicitConfig(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "lint.yaml"), strings.TrimSpace(`
commands:
  - name: "shell check"
    cmd: "sh"
    args: ["-c", "true"]
max_lines_by_ext:
  .go: 10
include:
  - "*.go"
`)+"\n")
	writeFile(t, filepath.Join(projectRoot, "ok.go"), "package main\n\nfunc main() {}\n")

	result := runForgekit(t, "--project-root", projectRoot, "lint", "--config", "lint.yaml")
	if result.err != nil {
		t.Fatalf("forgekit lint failed: %v\noutput:\n%s", result.err, result.output)
	}

	if !strings.Contains(result.output, "passed") {
		t.Fatalf("expected success markers in output, got:\n%s", result.output)
	}
}

func TestForgekitLintFailsWhenLineLimitExceeded(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "lint.yaml"), strings.TrimSpace(`
max_lines_by_ext:
  .go: 2
include:
  - "*.go"
`)+"\n")
	writeFile(t, filepath.Join(projectRoot, "too_long.go"), "package main\n\nfunc main() {}\n")

	result := runForgekit(t, "--project-root", projectRoot, "lint", "--config", "lint.yaml")
	if result.err == nil {
		t.Fatal("expected lint failure, got success")
	}

	if !strings.Contains(result.output, "some files exceed line limits") {
		t.Fatalf("unexpected output:\n%s", result.output)
	}
}

func TestForgekitPublishChartBuildFailsWhenMultiTagWithoutSemver(t *testing.T) {
	projectRoot := t.TempDir()
	t.Setenv("CHART_REGISTRY", "registry.example/charts")

	result := runForgekit(t,
		"--project-root", projectRoot,
		"publish", "chart", "build",
		"--chart-dir", "operator/chart",
		"--push",
		"--multi-tag",
	)

	if result.err == nil {
		t.Fatal("expected publish chart build to fail, got success")
	}

	if !strings.Contains(result.output, "--multi-tag requires --semver") {
		t.Fatalf("unexpected output:\n%s", result.output)
	}
}

type commandResult struct {
	output string
	err    error
}

func runForgekit(t *testing.T, args ...string) commandResult {
	t.Helper()

	root := repositoryRoot(t)
	commandArgs := append([]string{"run", "./cmd/forgekit"}, args...)
	cmd := exec.Command("go", commandArgs...)
	cmd.Dir = root
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	return commandResult{output: string(output), err: err}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
