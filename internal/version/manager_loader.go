package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func newManager(projectRoot string) (*Manager, error) {
	configPath := filepath.Join(projectRoot, "version-control.yaml")
	controlConfig, err := loadVersionControlConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load version-control.yaml: %w", err)
	}

	binaries, err := buildBinaryInfos(projectRoot, controlConfig.Binaries)
	if err != nil {
		return nil, err
	}

	charts, err := buildChartInfos(projectRoot, controlConfig.Charts)
	if err != nil {
		return nil, err
	}

	return &Manager{ProjectRoot: projectRoot, Charts: charts, Binaries: binaries}, nil
}

func buildBinaryInfos(projectRoot string, binaries []binaryControlConfig) ([]BinaryInfo, error) {
	items := make([]BinaryInfo, 0, len(binaries))
	names := make(map[string]struct{}, len(binaries))

	for _, binaryConfig := range binaries {
		name := strings.TrimSpace(binaryConfig.Name)
		if name == "" {
			return nil, fmt.Errorf("binary name is required")
		}

		binaryPath := strings.TrimSpace(binaryConfig.Path)
		if binaryPath == "" {
			return nil, fmt.Errorf("binary path is required for %s", name)
		}

		if _, exists := names[name]; exists {
			return nil, fmt.Errorf("duplicate binary name: %s", name)
		}
		names[name] = struct{}{}

		versionFile := strings.TrimSpace(binaryConfig.VersionFile)
		if versionFile == "" {
			versionFile = "VERSION"
		}

		basePath := filepath.Join(projectRoot, binaryPath)
		versionPath := filepath.Join(basePath, versionFile)

		items = append(items, BinaryInfo{
			Name:        name,
			BasePath:    basePath,
			VersionFile: versionFile,
			VersionPath: versionPath,
		})
	}

	return items, nil
}

func buildChartInfos(projectRoot string, charts []chartControlConfig) ([]ChartInfo, error) {
	items := make([]ChartInfo, 0, len(charts))

	for _, chartConfig := range charts {
		chartPath := filepath.Join(projectRoot, chartConfig.Path, "Chart.yaml")
		valuesPath := filepath.Join(projectRoot, chartConfig.Path, "values.yaml")

		images, err := extractImagesFromChart(chartPath, chartConfig.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from %s: %w", chartPath, err)
		}

		items = append(items, ChartInfo{
			Name:       chartConfig.Name,
			ChartPath:  chartPath,
			ValuesPath: valuesPath,
			Images:     images,
		})
	}

	return items, nil
}

func loadVersionControlConfig(path string) (*versionControlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config versionControlConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
