package version_test

import (
	"path/filepath"
	"strings"
	"testing"

	versioncmd "github.com/ben-wangz/forgekit/internal/version"
)

func TestBinaryConfigValidation(t *testing.T) {
	t.Run("missing binary name", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "version-control.yaml"), "binaries:\n  - path: .\n    versionFile: VERSION\n")

		_, err := versioncmd.NewManager(root)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "binary name is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing binary path", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "version-control.yaml"), "binaries:\n  - name: forgekit\n    versionFile: VERSION\n")

		_, err := versioncmd.NewManager(root)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "binary path is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate binary name", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "version-control.yaml"), "binaries:\n  - name: forgekit\n    path: release/a\n  - name: forgekit\n    path: release/b\n")

		_, err := versioncmd.NewManager(root)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate binary name") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestBinaryVersionFileDefaultsToVERSION(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "version-control.yaml"), "binaries:\n  - name: forgekit\n    path: release/forgekit\n")
	writeFile(t, filepath.Join(root, "release/forgekit/VERSION"), "0.3.1\n")

	manager, err := versioncmd.NewManager(root)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	version, err := manager.ModuleVersion("forgekit")
	if err != nil {
		t.Fatalf("failed to read binary version: %v", err)
	}

	if version != "0.3.1" {
		t.Fatalf("unexpected version: %s", version)
	}
}
