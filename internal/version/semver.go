package version

import (
	"fmt"
	"strconv"
	"strings"
)

func parseVersion(version string) (int, int, int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s (expected X.Y.Z)", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

func bumpVersion(version string, bumpType string) (string, error) {
	major, minor, patch, err := parseVersion(version)
	if err != nil {
		return "", err
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid bump type: %s (must be major, minor, or patch)", bumpType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
