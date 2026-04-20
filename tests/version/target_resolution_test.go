package version_test

import (
	"path/filepath"
	"strings"
	"testing"

	versioncmd "github.com/ben-wangz/forgekit/internal/version"
)

func TestVersionTargetResolutionPriority(t *testing.T) {
	root := t.TempDir()

	config := strings.Join([]string{
		"charts:",
		"  - name: demo-chart",
		"    path: operator/chart",
		"binaries:",
		"  - name: app",
		"    path: release/app",
		"    versionFile: VERSION",
	}, "\n") + "\n"

	chartYAML := strings.Join([]string{
		"apiVersion: v2",
		"name: demo-chart",
		"version: 0.1.0",
		"annotations:",
		"  demo-chart/images: |",
		"    - name: app",
		"      path: images/app",
		"      valuesKey: image.tag",
		"    - name: worker",
		"      path: images/worker",
		"      valuesKey: worker.image.tag",
	}, "\n") + "\n"

	writeFile(t, filepath.Join(root, "version-control.yaml"), config)
	writeFile(t, filepath.Join(root, "operator/chart/Chart.yaml"), chartYAML)
	writeFile(t, filepath.Join(root, "release/app/VERSION"), "1.0.0\n")
	writeFile(t, filepath.Join(root, "images/app/VERSION"), "2.0.0\n")
	writeFile(t, filepath.Join(root, "images/worker/VERSION"), "3.0.0\n")
	writeFile(t, filepath.Join(root, "legacy/container/VERSION"), "4.0.0\n")

	manager, err := versioncmd.NewManager(root)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	binaryPath := filepath.Join(root, "release/app/VERSION")
	if got := manager.VersionFilePath("app"); got != binaryPath {
		t.Fatalf("binary path should have highest priority, got %s", got)
	}

	if version, err := manager.ModuleVersion("app"); err != nil || version != "1.0.0" {
		t.Fatalf("unexpected app version: version=%s err=%v", version, err)
	}

	imagePath := filepath.Join(root, "images/worker/VERSION")
	if got := manager.VersionFilePath("worker"); got != imagePath {
		t.Fatalf("unexpected worker path: %s", got)
	}

	if version, err := manager.ModuleVersion("worker"); err != nil || version != "3.0.0" {
		t.Fatalf("unexpected worker version: version=%s err=%v", version, err)
	}

	legacyPath := filepath.Join(root, "legacy/container/VERSION")
	if got := manager.VersionFilePath("legacy"); got != legacyPath {
		t.Fatalf("unexpected legacy fallback path: %s", got)
	}

	if version, err := manager.ModuleVersion("legacy"); err != nil || version != "4.0.0" {
		t.Fatalf("unexpected legacy version: version=%s err=%v", version, err)
	}
}
