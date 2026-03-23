package project

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolveRoot(explicit string) (string, error) {
	candidate := explicit
	if candidate == "" {
		candidate = os.Getenv("FORGEKIT_PROJECT_ROOT")
	}

	if candidate != "" {
		return validateRoot(candidate)
	}

	start, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to read current directory: %w", err)
	}

	root, err := discoverRoot(start)
	if err != nil {
		return "", fmt.Errorf("%w (set --project-root or FORGEKIT_PROJECT_ROOT)", err)
	}

	return root, nil
}

func ResolvePath(projectRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}

func validateRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %q: %w", path, err)
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to access project root %q: %w", absPath, err)
	}

	if !stat.IsDir() {
		return "", fmt.Errorf("project root %q is not a directory", absPath)
	}

	return absPath, nil
}

func discoverRoot(start string) (string, error) {
	dir := start

	for {
		if hasMarker(dir, ".git") || hasMarker(dir, "version-control.yaml") {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parent
	}
}

func hasMarker(dir, marker string) bool {
	_, err := os.Stat(filepath.Join(dir, marker))
	return err == nil
}
