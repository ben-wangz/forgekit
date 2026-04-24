package version

import (
	"fmt"
	"path/filepath"
)

func versionFilePath(m *Manager, module string) string {
	for _, binary := range m.Binaries {
		if binary.Name == module {
			return binary.VersionPath
		}
	}

	for _, chart := range m.Charts {
		for _, image := range chart.Images {
			if image.Name == module {
				return filepath.Join(m.ProjectRoot, image.Path, "VERSION")
			}
		}
	}

	return filepath.Join(m.ProjectRoot, module, "container", "VERSION")
}

func findImageByName(m *Manager, name string) (*ImageConfig, *ChartInfo) {
	for i := range m.Charts {
		for j := range m.Charts[i].Images {
			if m.Charts[i].Images[j].Name == name {
				return &m.Charts[i].Images[j], &m.Charts[i]
			}
		}
	}

	return nil, nil
}

func chartByName(m *Manager, name string) (*ChartInfo, error) {
	for i := range m.Charts {
		if m.Charts[i].Name == name {
			return &m.Charts[i], nil
		}
	}

	return nil, fmt.Errorf("chart not found: %s", name)
}
