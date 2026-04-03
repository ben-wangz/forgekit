package publish

import (
	"fmt"
	"strconv"
	"strings"
)

func resolvePublishTags(version string, semver bool, push bool, multiTag bool) ([]string, []string, error) {
	if !multiTag {
		return []string{version}, nil, nil
	}

	if !semver {
		return nil, nil, fmt.Errorf("--multi-tag requires --semver")
	}

	if !push {
		return nil, nil, fmt.Errorf("--multi-tag requires --push")
	}

	major, minor, hasPrerelease, err := parseSemverForTags(version)
	if err != nil {
		return nil, nil, err
	}

	if hasPrerelease {
		warning := "[WARN] prerelease version detected, --multi-tag is downgraded to single tag publish"
		return []string{version}, []string{warning}, nil
	}

	if major == 0 {
		warning := "[WARN] pre-1.0 semver detected, --multi-tag is downgraded to single tag publish"
		return []string{version}, []string{warning}, nil
	}

	tags := []string{
		"latest",
		strconv.Itoa(major),
		fmt.Sprintf("%d.%d", major, minor),
		version,
	}

	return dedupeStrings(tags), nil, nil
}

func parseSemverForTags(version string) (int, int, bool, error) {
	if strings.Contains(version, "+") {
		return 0, 0, false, fmt.Errorf("OCI tag does not support '+': %s", version)
	}

	hasPrerelease := false
	core := version
	if idx := strings.Index(version, "-"); idx >= 0 {
		hasPrerelease = true
		core = version[:idx]
	}

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return 0, 0, false, fmt.Errorf("--multi-tag requires stable semver format X.Y.Z, got: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid semver major in %s", version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid semver minor in %s", version)
	}

	if _, err := strconv.Atoi(parts[2]); err != nil {
		return 0, 0, false, fmt.Errorf("invalid semver patch in %s", version)
	}

	return major, minor, hasPrerelease, nil
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
