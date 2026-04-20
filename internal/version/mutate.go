package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdBump(manager *Manager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: forgekit version bump <target> <major|minor|patch>")
	}

	target := args[0]
	bumpType := args[1]

	return bumpImageVersion(manager, target, bumpType)
}

func cmdBumpChart(manager *Manager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: forgekit version bump-chart <chart-name> <major|minor|patch> [--sync]")
	}

	syncImages := false
	plainArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--sync" {
			syncImages = true
			continue
		}
		plainArgs = append(plainArgs, arg)
	}

	if len(plainArgs) == 0 {
		return fmt.Errorf("usage: forgekit version bump-chart <chart-name> <major|minor|patch> [--sync]")
	}

	if len(plainArgs) == 1 {
		if len(manager.Charts) != 1 {
			return fmt.Errorf("chart name is required when there are multiple charts")
		}
		return bumpChartVersion(manager, manager.Charts[0].Name, plainArgs[0], syncImages)
	}

	return bumpChartVersion(manager, plainArgs[0], plainArgs[1], syncImages)
}

func cmdSync(manager *Manager, args []string) error {
	if len(args) > 0 {
		chartName := args[0]
		chart, err := manager.chartByName(chartName)
		if err != nil {
			return err
		}
		return syncImagesToChart(manager, chart)
	}

	for i := range manager.Charts {
		fmt.Printf("Syncing chart: %s\n", manager.Charts[i].Name)
		if err := syncImagesToChart(manager, &manager.Charts[i]); err != nil {
			return err
		}
		fmt.Println()
	}

	return nil
}

func bumpImageVersion(manager *Manager, module, bumpType string) error {
	versionFile := manager.VersionFilePath(module)

	if _, err := os.Stat(versionFile); err != nil {
		return fmt.Errorf("version file not found: %s", versionFile)
	}

	data, err := os.ReadFile(versionFile)
	if err != nil {
		return err
	}

	currentVersion := strings.TrimSpace(string(data))
	fmt.Printf("Current version: %s\n", currentVersion)

	newVersion, err := bumpVersion(currentVersion, bumpType)
	if err != nil {
		return err
	}

	fmt.Printf("New version: %s\n", newVersion)

	if err := os.WriteFile(versionFile, []byte(newVersion+"\n"), 0644); err != nil {
		return err
	}

	fmt.Printf("\n✓ Version bump complete: %s -> %s\n", currentVersion, newVersion)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the changes: git diff")
	fmt.Println("  2. Update chart if needed: forgekit version bump-chart <chart-name> <type> --sync")
	fmt.Println("  3. Commit the changes")

	return nil
}

func bumpChartVersion(manager *Manager, chartName, bumpType string, syncImages bool) error {
	targetChart, err := manager.chartByName(chartName)
	if err != nil {
		return err
	}

	currentVersion, err := readYAMLValue(targetChart.ChartPath, "version")
	if err != nil {
		return err
	}

	fmt.Printf("Current chart version: %s\n", currentVersion)

	newVersion, err := bumpVersion(currentVersion, bumpType)
	if err != nil {
		return err
	}

	fmt.Printf("New chart version: %s\n", newVersion)

	if syncImages {
		fmt.Println("\nSyncing image versions...")
		if err := syncImagesToChart(manager, targetChart); err != nil {
			return err
		}
	}

	if err := updateYAMLValue(targetChart.ChartPath, "version", newVersion); err != nil {
		return err
	}

	fmt.Printf("\n✓ Chart version bump complete: %s -> %s\n", currentVersion, newVersion)
	if syncImages {
		fmt.Println("✓ Image versions synced")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the changes: git diff")
	fmt.Println("  2. Commit the changes")

	return nil
}

func syncImagesToChart(manager *Manager, chart *ChartInfo) error {
	if len(chart.Images) == 0 {
		fmt.Printf("  No images configured for chart %s\n", chart.Name)
		return nil
	}

	firstImageVersion := ""

	for _, image := range chart.Images {
		versionFile := filepath.Join(manager.ProjectRoot, image.Path, "VERSION")
		data, err := os.ReadFile(versionFile)
		if err != nil {
			fmt.Printf("  Warning: VERSION file not found for %s, skipping\n", image.Name)
			continue
		}

		version := strings.TrimSpace(string(data))
		if firstImageVersion == "" {
			firstImageVersion = version
		}

		if err := updateYAMLValue(chart.ValuesPath, image.ValuesKey, version); err != nil {
			return fmt.Errorf("failed to update %s in values.yaml: %w", image.ValuesKey, err)
		}

		fmt.Printf("  ✓ Updated %s to %s\n", image.ValuesKey, version)
	}

	if firstImageVersion != "" {
		if err := updateYAMLValue(chart.ChartPath, "appVersion", firstImageVersion); err != nil {
			return fmt.Errorf("failed to update appVersion: %w", err)
		}
		fmt.Printf("  ✓ Updated appVersion to %s\n", firstImageVersion)
	}

	return nil
}
