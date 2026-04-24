package version

import (
	"fmt"

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
