package publish_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	publish "github.com/ben-wangz/forgekit/internal/publish"
)

func TestChartBuildMultiTagRequiresSemver(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHART_REGISTRY", "registry.example/charts")

	err := publish.Run([]string{"chart", "build", "--chart-dir", "operator/chart", "--push", "--multi-tag"}, root)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "--semver") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestChartBuildMultiTagRequiresPush(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHART_REGISTRY", "registry.example/charts")

	err := publish.Run([]string{"chart", "build", "--chart-dir", "operator/chart", "--semver", "--multi-tag"}, root)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "--push") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestChartBuildMultiTagRejectsBuildMetadataBeforePackage(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CHART_REGISTRY", "registry.example/charts")
	versionControlPath := filepath.Join(root, "version-control.yaml")
	chartDir := filepath.Join(root, "operator", "chart")
	chartYAMLPath := filepath.Join(chartDir, "Chart.yaml")

	if err := os.WriteFile(versionControlPath, []byte("charts:\n  - name: demo\n    path: operator/chart\n"), 0644); err != nil {
		t.Fatalf("failed to write version-control.yaml: %v", err)
	}
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatalf("failed to create chart dir: %v", err)
	}
	chartYAML := "apiVersion: v2\nname: demo\nversion: 1.6.0+build.1\n"
	if err := os.WriteFile(chartYAMLPath, []byte(chartYAML), 0644); err != nil {
		t.Fatalf("failed to write Chart.yaml: %v", err)
	}

	if err := initGitRepo(root); err != nil {
		t.Fatalf("failed to initialize git repo: %v", err)
	}

	err := publish.Run([]string{"chart", "build", "--chart-dir", "operator/chart", "--push", "--semver", "--multi-tag"}, root)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "OCI tag") || !strings.Contains(err.Error(), "+") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func initGitRepo(root string) error {
	if err := runGit(root, "init"); err != nil {
		return err
	}
	if err := runGit(root, "add", "."); err != nil {
		return err
	}
	if err := runGit(root, "-c", "user.name=forgekit-test", "-c", "user.email=forgekit-test@example.com", "commit", "-m", "init"); err != nil {
		return err
	}
	return nil
}

func runGit(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	_ = output
	return nil
}
