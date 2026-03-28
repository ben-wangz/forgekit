package version

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateYAMLValue_PreservesCommentsAndFormatting(t *testing.T) {
	before := "# values for operator\nimage:\n  repository: docker.io/acme/operator  # keep repo comment\n\n  tag: \"0.1.0\" # keep inline comment\n\n# preserve this block\nextra:\n  enabled: true\n"
	after := "# values for operator\nimage:\n  repository: docker.io/acme/operator  # keep repo comment\n\n  tag: \"0.1.1\" # keep inline comment\n\n# preserve this block\nextra:\n  enabled: true\n"

	filePath := writeTempFile(t, "values.yaml", before)

	if err := updateYAMLValue(filePath, "image.tag", "0.1.1"); err != nil {
		t.Fatalf("updateYAMLValue failed: %v", err)
	}

	got := readFile(t, filePath)
	if got != after {
		t.Fatalf("unexpected content:\n--- expected ---\n%s\n--- actual ---\n%s", after, got)
	}
}

func TestUpdateYAMLValue_MultipleKeyPathsInSameFile(t *testing.T) {
	before := "# image tags\nimage:\n  tag: 0.1.0        # operator\n\nworkflow:\n  image:\n    tag: '0.2.0'    # workflow\n\nmcp:\n  image:\n    tag: \"0.3.0\"    # mcp\n"
	after := "# image tags\nimage:\n  tag: 1.1.0        # operator\n\nworkflow:\n  image:\n    tag: '1.2.0'    # workflow\n\nmcp:\n  image:\n    tag: \"1.3.0\"    # mcp\n"

	filePath := writeTempFile(t, "values.yaml", before)

	updates := []struct {
		keyPath string
		value   string
	}{
		{keyPath: "image.tag", value: "1.1.0"},
		{keyPath: "workflow.image.tag", value: "1.2.0"},
		{keyPath: "mcp.image.tag", value: "1.3.0"},
	}

	for _, update := range updates {
		if err := updateYAMLValue(filePath, update.keyPath, update.value); err != nil {
			t.Fatalf("updateYAMLValue failed for %s: %v", update.keyPath, err)
		}
	}

	got := readFile(t, filePath)
	if got != after {
		t.Fatalf("unexpected content:\n--- expected ---\n%s\n--- actual ---\n%s", after, got)
	}
}

func TestUpdateYAMLValue_ChartAppVersionPreservesOrderAndComments(t *testing.T) {
	before := "apiVersion: v2\nname: example-chart\n# chart version comment\nversion: 0.1.0\n# app version comment\nappVersion: \"0.1.0\"\n\ndependencies:\n  - name: redis\n    version: 18.0.0\n    repository: oci://registry.example/charts\n"
	after := "apiVersion: v2\nname: example-chart\n# chart version comment\nversion: 0.1.0\n# app version comment\nappVersion: \"0.2.0\"\n\ndependencies:\n  - name: redis\n    version: 18.0.0\n    repository: oci://registry.example/charts\n"

	filePath := writeTempFile(t, "Chart.yaml", before)

	if err := updateYAMLValue(filePath, "appVersion", "0.2.0"); err != nil {
		t.Fatalf("updateYAMLValue failed: %v", err)
	}

	got := readFile(t, filePath)
	if got != after {
		t.Fatalf("unexpected content:\n--- expected ---\n%s\n--- actual ---\n%s", after, got)
	}
}

func TestUpdateYAMLValue_ErrorCases(t *testing.T) {
	t.Run("keyPath not found", func(t *testing.T) {
		filePath := writeTempFile(t, "values.yaml", "image:\n  tag: 0.1.0\n")

		err := updateYAMLValue(filePath, "workflow.image.tag", "1.0.0")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "file=") || !strings.Contains(err.Error(), "keyPath=workflow.image.tag") {
			t.Fatalf("error message does not include filePath/keyPath details: %v", err)
		}
		if !strings.Contains(err.Error(), "keyPath not found") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("keyPath points to non-scalar", func(t *testing.T) {
		content := "image:\n  tag:\n    nested: value\n"
		filePath := writeTempFile(t, "values.yaml", content)

		err := updateYAMLValue(filePath, "image.tag", "1.0.0")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "target is not scalar") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func writeTempFile(t *testing.T, fileName, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return filePath
}

func readFile(t *testing.T, filePath string) string {
	t.Helper()

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	return string(data)
}
