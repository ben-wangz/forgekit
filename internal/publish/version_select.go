package publish

import (
	"fmt"

	versiontool "github.com/ben-wangz/forgekit/internal/version"
)

type publishVersion struct {
	Value string
	Mode  string
}

func resolveModulePublishVersion(manager *versiontool.Manager, module string, semver bool) (*publishVersion, error) {
	if semver {
		dirty, err := manager.GitDirty()
		if err != nil {
			return nil, fmt.Errorf("failed to check git dirty state: %w", err)
		}
		if dirty {
			return nil, fmt.Errorf("working tree is dirty; --semver requires a clean repository (use default git-version mode)")
		}

		version, err := manager.ModuleVersion(module)
		if err != nil {
			return nil, err
		}

		return &publishVersion{Value: version, Mode: "semver"}, nil
	}

	version, err := manager.ModuleGitVersion(module)
	if err != nil {
		return nil, err
	}

	return &publishVersion{Value: version, Mode: "git-version"}, nil
}

func resolveChartPublishVersion(manager *versiontool.Manager, chartName string, semver bool) (*publishVersion, error) {
	if semver {
		dirty, err := manager.GitDirty()
		if err != nil {
			return nil, fmt.Errorf("failed to check git dirty state: %w", err)
		}
		if dirty {
			return nil, fmt.Errorf("working tree is dirty; --semver requires a clean repository (use default git-version mode)")
		}

		version, err := manager.ChartVersion(chartName)
		if err != nil {
			return nil, err
		}

		return &publishVersion{Value: version, Mode: "semver"}, nil
	}

	version, err := manager.ChartGitVersion(chartName)
	if err != nil {
		return nil, err
	}

	return &publishVersion{Value: version, Mode: "git-version"}, nil
}
