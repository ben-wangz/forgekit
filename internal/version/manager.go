package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ImageConfig struct {
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
	ValuesKey string `yaml:"valuesKey"`
}

type binaryControlConfig struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	VersionFile string `yaml:"versionFile"`
}

type chartControlConfig struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type versionControlConfig struct {
	Charts   []chartControlConfig  `yaml:"charts"`
	Binaries []binaryControlConfig `yaml:"binaries"`
}

type ChartInfo struct {
	Name       string
	ChartPath  string
	ValuesPath string
	Images     []ImageConfig
}

type BinaryInfo struct {
	Name        string
	BasePath    string
	VersionFile string
	VersionPath string
}

type Manager struct {
	ProjectRoot string
	Charts      []ChartInfo
	Binaries    []BinaryInfo
}

func NewManager(projectRoot string) (*Manager, error) {
	configPath := filepath.Join(projectRoot, "version-control.yaml")
	controlConfig, err := loadVersionControlConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load version-control.yaml: %w", err)
	}

	charts := make([]ChartInfo, 0, len(controlConfig.Charts))
	binaries := make([]BinaryInfo, 0, len(controlConfig.Binaries))
	binaryNames := make(map[string]struct{}, len(controlConfig.Binaries))

	for _, binaryConfig := range controlConfig.Binaries {
		name := strings.TrimSpace(binaryConfig.Name)
		if name == "" {
			return nil, fmt.Errorf("binary name is required")
		}

		binaryPath := strings.TrimSpace(binaryConfig.Path)
		if binaryPath == "" {
			return nil, fmt.Errorf("binary path is required for %s", name)
		}

		if _, exists := binaryNames[name]; exists {
			return nil, fmt.Errorf("duplicate binary name: %s", name)
		}
		binaryNames[name] = struct{}{}

		versionFile := strings.TrimSpace(binaryConfig.VersionFile)
		if versionFile == "" {
			versionFile = "VERSION"
		}

		basePath := filepath.Join(projectRoot, binaryPath)
		versionPath := filepath.Join(basePath, versionFile)

		binaries = append(binaries, BinaryInfo{
			Name:        name,
			BasePath:    basePath,
			VersionFile: versionFile,
			VersionPath: versionPath,
		})
	}

	for _, chartConfig := range controlConfig.Charts {
		chartPath := filepath.Join(projectRoot, chartConfig.Path, "Chart.yaml")
		valuesPath := filepath.Join(projectRoot, chartConfig.Path, "values.yaml")

		images, err := extractImagesFromChart(chartPath, chartConfig.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from %s: %w", chartPath, err)
		}

		charts = append(charts, ChartInfo{
			Name:       chartConfig.Name,
			ChartPath:  chartPath,
			ValuesPath: valuesPath,
			Images:     images,
		})
	}

	return &Manager{ProjectRoot: projectRoot, Charts: charts, Binaries: binaries}, nil
}

func (m *Manager) VersionFilePath(module string) string {
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

func (m *Manager) findImageByName(name string) (*ImageConfig, *ChartInfo) {
	for i := range m.Charts {
		for j := range m.Charts[i].Images {
			if m.Charts[i].Images[j].Name == name {
				return &m.Charts[i].Images[j], &m.Charts[i]
			}
		}
	}

	return nil, nil
}

func (m *Manager) chartByName(name string) (*ChartInfo, error) {
	for i := range m.Charts {
		if m.Charts[i].Name == name {
			return &m.Charts[i], nil
		}
	}

	return nil, fmt.Errorf("chart not found: %s", name)
}

func (m *Manager) ModuleVersion(module string) (string, error) {
	versionFile := m.VersionFilePath(module)
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", fmt.Errorf("version file not found: %s", versionFile)
	}

	return strings.TrimSpace(string(data)), nil
}

func (m *Manager) ChartVersion(chartName string) (string, error) {
	chart, err := m.chartByName(chartName)
	if err != nil {
		return "", err
	}

	return readYAMLValue(chart.ChartPath, "version")
}

func (m *Manager) AppVersion(chartName string) (string, error) {
	chart, err := m.chartByName(chartName)
	if err != nil {
		return "", err
	}

	version, err := readYAMLValue(chart.ChartPath, "appVersion")
	if err != nil {
		return "", err
	}

	return strings.Trim(version, "\""), nil
}

func (m *Manager) ModuleGitVersion(module string) (string, error) {
	version, err := m.ModuleVersion(module)
	if err != nil {
		return "", err
	}

	commitID, err := getGitCommitID(m.ProjectRoot)
	if err != nil {
		return "", err
	}

	dirty, err := isGitDirty(m.ProjectRoot)
	if err != nil {
		return "", err
	}

	return buildGitVersion(version, commitID, dirty), nil
}

func (m *Manager) ChartGitVersion(chartName string) (string, error) {
	version, err := m.ChartVersion(chartName)
	if err != nil {
		return "", err
	}

	commitID, err := getGitCommitID(m.ProjectRoot)
	if err != nil {
		return "", err
	}

	dirty, err := isGitDirty(m.ProjectRoot)
	if err != nil {
		return "", err
	}

	return buildGitVersion(version, commitID, dirty), nil
}

func (m *Manager) GitDirty() (bool, error) {
	return isGitDirty(m.ProjectRoot)
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

func extractImagesFromChart(chartPath, chartName string) ([]ImageConfig, error) {
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart map[string]any
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	annotations, ok := chart["annotations"].(map[string]any)
	if !ok {
		return []ImageConfig{}, nil
	}

	annotationKey := fmt.Sprintf("%s/images", chartName)
	imagesYAML, ok := annotations[annotationKey].(string)
	if !ok || imagesYAML == "" {
		return []ImageConfig{}, nil
	}

	var images []ImageConfig
	if err := yaml.Unmarshal([]byte(imagesYAML), &images); err != nil {
		return nil, fmt.Errorf("failed to parse images annotation: %w", err)
	}

	return images, nil
}
