package publish

import (
	"fmt"
	"os"
	"path/filepath"

	versiontool "github.com/ben-wangz/forgekit/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	deprecatedRegistryWarning   = "[WARN] DEPRECATED: chart registry is resolved via compatibility fallback path. Please set CHART_REGISTRY explicitly. This compatibility may be removed in the next release."
	deprecatedCredentialWarning = "[WARN] DEPRECATED: chart credentials are resolved via compatibility fallback path. Please set CHART_REGISTRY_USERNAME and CHART_REGISTRY_PASSWORD explicitly. This compatibility may be removed in the next release."
)

type ChartYAML struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func cmdChartBuild(args []string, projectRoot string) error {
	cfg, err := loadChartConfig(args, projectRoot)
	if err != nil {
		return err
	}

	if cfg.RegistryFallbackDeprecated {
		fmt.Fprintln(os.Stderr, deprecatedRegistryWarning)
	}
	if cfg.CredentialFallbackDeprecated {
		fmt.Fprintln(os.Stderr, deprecatedCredentialWarning)
	}

	chartYAML, err := readChartYAML(cfg.ChartDir)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	versionManager, err := versiontool.NewManager(cfg.ProjectRoot)
	if err != nil {
		return err
	}

	publishVersion, err := resolveChartPublishVersion(versionManager, chartYAML.Name, cfg.Semver)
	if err != nil {
		return fmt.Errorf("failed to resolve chart publish version: %w", err)
	}

	chartVersion := publishVersion.Value

	printChartBuildInfo(cfg, chartYAML.Name, publishVersion.Mode, chartYAML.Version, chartVersion)

	fmt.Println()
	fmt.Println("Packaging chart...")
	fmt.Println()

	chartPackage, err := helmPackage(cfg.ChartDir, chartVersion)
	if err != nil {
		return fmt.Errorf("failed to package chart: %w", err)
	}
	defer os.Remove(chartPackage)

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Package Complete")
	fmt.Println("===========================================")
	fmt.Printf("Chart package: %s\n", chartPackage)
	fmt.Println()

	if cfg.Push {
		if err := pushChart(cfg, chartPackage, chartYAML.Name, chartVersion); err != nil {
			return err
		}
	}

	return nil
}

func readChartYAML(chartDir string) (*ChartYAML, error) {
	chartFile := filepath.Join(chartDir, "Chart.yaml")
	data, err := os.ReadFile(chartFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart ChartYAML
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	return &chart, nil
}

func printChartBuildInfo(cfg *ChartConfig, chartName, versionMode, baseVersion, fullVersion string) {
	fmt.Println("===========================================")
	fmt.Println("Building Helm Chart")
	fmt.Println("===========================================")
	fmt.Printf("Project root: %s\n", cfg.ProjectRoot)
	fmt.Printf("Chart dir: %s\n", cfg.ChartDir)
	fmt.Printf("Chart name: %s\n", chartName)
	fmt.Printf("Version mode: %s\n", versionMode)
	fmt.Printf("Base version: %s\n", baseVersion)
	fmt.Printf("Package version: %s\n", fullVersion)
	if cfg.Push {
		fmt.Println("Push: enabled")
		fmt.Printf("Registry: %s\n", cfg.ChartRegistry)
	}
}

func pushChart(cfg *ChartConfig, chartPackage, chartName, chartVersion string) error {
	fmt.Println("===========================================")
	fmt.Println("Pushing Helm Chart")
	fmt.Println("===========================================")

	registryPath := cfg.ChartRegistry
	chartRef := fmt.Sprintf("%s/%s:%s", cfg.ChartRegistry, chartName, chartVersion)
	fmt.Printf("Chart reference: %s\n", chartRef)
	fmt.Println()

	if cfg.Username != "" && cfg.Password != "" {
		fmt.Println("Logging in to registry...")
		if err := helmRegistryLogin(cfg.ChartRegistryHost, cfg.Username, cfg.Password, cfg.RegistryPlainHTTP); err != nil {
			return fmt.Errorf("failed to login to registry: %w", err)
		}
		fmt.Println()
	}

	fmt.Println("Pushing chart...")
	if err := helmPush(chartPackage, registryPath, cfg.RegistryPlainHTTP); err != nil {
		return fmt.Errorf("failed to push chart: %w", err)
	}

	if cfg.Username != "" && cfg.Password != "" {
		fmt.Println()
		fmt.Println("Logging out from registry...")
		if err := helmRegistryLogout(cfg.ChartRegistryHost); err != nil {
			return fmt.Errorf("failed to logout from registry: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Push Complete")
	fmt.Println("===========================================")
	fmt.Printf("Chart: %s\n", chartRef)
	fmt.Println()

	return nil
}
