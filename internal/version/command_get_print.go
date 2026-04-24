package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func printAllVersions(manager *Manager) error {
	if len(manager.Binaries) > 0 {
		fmt.Println("=== Binaries ===")
		fmt.Println()

		for _, binary := range manager.Binaries {
			version, err := manager.ModuleVersion(binary.Name)
			if err != nil {
				return err
			}

			fmt.Printf("  %-30s version: %s\n", binary.Name, version)
		}

		fmt.Println()
	}

	fmt.Println("=== Helm Charts ===")
	fmt.Println()

	for _, chart := range manager.Charts {
		chartVersion, err := readYAMLValue(chart.ChartPath, "version")
		if err != nil {
			return err
		}

		appVersion, err := readYAMLValue(chart.ChartPath, "appVersion")
		if err != nil {
			appVersion = "N/A"
		}
		appVersion = strings.Trim(appVersion, "\"")

		fmt.Printf("  %-30s chart: %-10s appVersion: %s\n", chart.Name, chartVersion, appVersion)

		if len(chart.Images) > 0 {
			fmt.Println()
			for _, image := range chart.Images {
				versionFile := filepath.Join(manager.ProjectRoot, image.Path, "VERSION")
				if data, err := os.ReadFile(versionFile); err == nil {
					version := strings.TrimSpace(string(data))
					fmt.Printf("    %-28s image: %s\n", image.Name, version)
				}
			}
		}

		fmt.Println()
	}

	return nil
}

func printChartVersion(manager *Manager, args []string) error {
	chartName := ""
	if len(args) > 0 {
		chartName = args[0]
	}

	if chartName == "" {
		for _, chart := range manager.Charts {
			version, err := manager.ChartVersion(chart.Name)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", chart.Name, version)
		}
		return nil
	}

	version, err := manager.ChartVersion(chartName)
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}

func printChartGitVersion(manager *Manager, args []string) error {
	chartName := ""
	if len(args) > 0 {
		chartName = args[0]
	}

	if chartName == "" {
		for _, chart := range manager.Charts {
			version, err := manager.ChartGitVersion(chart.Name)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", chart.Name, version)
		}
		return nil
	}

	version, err := manager.ChartGitVersion(chartName)
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}

func printAppVersion(manager *Manager, args []string) error {
	chartName := ""
	if len(args) > 0 {
		chartName = args[0]
	}

	if chartName == "" {
		for _, chart := range manager.Charts {
			version, err := manager.AppVersion(chart.Name)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", chart.Name, version)
		}
		return nil
	}

	version, err := manager.AppVersion(chartName)
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}

func printModuleVersion(manager *Manager, module string) error {
	version, err := manager.ModuleVersion(module)
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}

func printModuleGitVersion(manager *Manager, module string) error {
	version, err := manager.ModuleGitVersion(module)
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}
