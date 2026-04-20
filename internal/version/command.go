package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ben-wangz/forgekit/internal/project"
)

func Run(args []string, defaultProjectRoot string) error {
	args, rootOverride, err := extractProjectRootFlag(args)
	if err != nil {
		return err
	}

	if len(args) < 1 {
		printUsage()
		return nil
	}

	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return nil
	}

	projectRoot := defaultProjectRoot
	if rootOverride != "" {
		projectRoot = rootOverride
	}

	resolvedRoot, err := project.ResolveRoot(projectRoot)
	if err != nil {
		return err
	}

	manager, err := NewManager(resolvedRoot)
	if err != nil {
		return err
	}

	cmd := args[0]
	subArgs := args[1:]

	switch cmd {
	case "get":
		return cmdGet(manager, subArgs)
	case "bump":
		return cmdBump(manager, subArgs)
	case "bump-chart":
		return cmdBumpChart(manager, subArgs)
	case "sync":
		return cmdSync(manager, subArgs)
	default:
		printUsage()
		return fmt.Errorf("unknown version command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Usage: forgekit version <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  get [target] [--git]                 List/get semantic version or git-version")
	fmt.Println("  bump <target> <major|minor|patch>    Bump target version")
	fmt.Println("  bump-chart <chart> <type> [--sync]   Bump chart version and optionally sync image versions")
	fmt.Println("  bump-chart <type> [--sync]           Bump the only chart when there is exactly one chart")
	fmt.Println("  sync [chart-name]                    Sync image versions to values.yaml and appVersion")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --project-root <path>                Override project root for this command")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  forgekit version get")
	fmt.Println("  forgekit version get forgekit")
	fmt.Println("  forgekit version get catalog/ingest")
	fmt.Println("  forgekit version get chart astro-data-operator --git")
	fmt.Println("  forgekit version bump forgekit patch")
	fmt.Println("  forgekit version bump catalog/ingest patch")
	fmt.Println("  forgekit version bump-chart astro-data-operator minor --sync")
	fmt.Println("  forgekit version sync")
}

func cmdGet(manager *Manager, args []string) error {
	gitVersion := false
	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--git" {
			gitVersion = true
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	if len(filteredArgs) == 0 {
		if gitVersion {
			return fmt.Errorf("usage: forgekit version get <module> --git")
		}
		return printAllVersions(manager)
	}

	module := filteredArgs[0]
	if gitVersion {
		if module == "chart" {
			if len(filteredArgs) > 1 && (filteredArgs[1] == "app" || filteredArgs[1] == "appVersion") {
				return fmt.Errorf("--git is not supported for chart appVersion")
			}
			return printChartGitVersion(manager, filteredArgs[1:])
		}
		return printModuleGitVersion(manager, module)
	}

	if module == "chart" {
		if len(filteredArgs) > 1 && (filteredArgs[1] == "app" || filteredArgs[1] == "appVersion") {
			return printAppVersion(manager, filteredArgs[2:])
		}
		return printChartVersion(manager, filteredArgs[1:])
	}

	return printModuleVersion(manager, module)
}

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

func extractProjectRootFlag(args []string) ([]string, string, error) {
	filtered := make([]string, 0, len(args))
	projectRoot := ""

	for i := 0; i < len(args); i++ {
		if args[i] == "--project-root" {
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("--project-root requires a value")
			}
			projectRoot = args[i+1]
			i++
			continue
		}
		filtered = append(filtered, args[i])
	}

	return filtered, projectRoot, nil
}
