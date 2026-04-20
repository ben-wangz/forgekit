package version_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	versioncmd "github.com/ben-wangz/forgekit/internal/version"
)

func TestVersionGetAndBumpBinary(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "version-control.yaml"), "binaries:\n  - name: forgekit\n    path: .\n    versionFile: VERSION\n")
	writeFile(t, filepath.Join(root, "VERSION"), "1.2.3\n")

	output, err := runVersionCommandWithOutput([]string{"get", "forgekit"}, root)
	if err != nil {
		t.Fatalf("version get failed: %v", err)
	}

	if strings.TrimSpace(output) != "1.2.3" {
		t.Fatalf("unexpected version get output: %q", output)
	}

	if err := versioncmd.Run([]string{"bump", "forgekit", "patch"}, root); err != nil {
		t.Fatalf("version bump failed: %v", err)
	}

	bumped := strings.TrimSpace(readFile(t, filepath.Join(root, "VERSION")))
	if bumped != "1.2.4" {
		t.Fatalf("unexpected bumped version: %s", bumped)
	}
}

func runVersionCommandWithOutput(args []string, root string) (string, error) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w
	runErr := versioncmd.Run(args, root)
	_ = w.Close()
	os.Stdout = originalStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	return buf.String(), runErr
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

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	return string(data)
}
