package version

import (
	"fmt"
	"os"
	"strings"
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
	return newManager(projectRoot)
}

func (m *Manager) VersionFilePath(module string) string {
	return versionFilePath(m, module)
}

func (m *Manager) findImageByName(name string) (*ImageConfig, *ChartInfo) {
	return findImageByName(m, name)
}

func (m *Manager) chartByName(name string) (*ChartInfo, error) {
	return chartByName(m, name)
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
