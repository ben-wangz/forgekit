package publish

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

	if len(args) < 2 {
		printUsage()
		return nil
	}

	category := args[0]
	cmd := args[1]
	subArgs := args[2:]

	switch category {
	case "container":
		return handleContainerCommand(cmd, subArgs, resolvedRoot)
	case "chart":
		return handleChartCommand(cmd, subArgs, resolvedRoot)
	default:
		printUsage()
		return fmt.Errorf("unknown category: %s", category)
	}
}

func handleContainerCommand(cmd string, args []string, projectRoot string) error {
	switch cmd {
	case "build":
		return cmdContainerBuild(args, projectRoot)
	default:
		return fmt.Errorf("unknown container command: %s", cmd)
	}
}

func handleChartCommand(cmd string, args []string, projectRoot string) error {
	switch cmd {
	case "build":
		return cmdChartBuild(args, projectRoot)
	default:
		return fmt.Errorf("unknown chart command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println("Usage: forgekit publish <category> <command> [options]")
	fmt.Println()
	fmt.Println("Categories:")
	fmt.Println("  container    Container image operations")
	fmt.Println("  chart        Helm chart operations")
	fmt.Println()
	fmt.Println("Container commands:")
	fmt.Println("  build        Build and optionally push container image")
	fmt.Println()
	fmt.Println("Container build options:")
	fmt.Println("  --container-dir <path>  Container directory path (required)")
	fmt.Println("  --module <name>         Module name for version logic (required)")
	fmt.Println("  --push                  Push image to registry")
	fmt.Println("  --load-k3s              Load image to k3s after build")
	fmt.Println("  --context <path>        Build context path (default: project root)")
	fmt.Println()
	fmt.Println("Chart commands:")
	fmt.Println("  build        Build and optionally push helm chart")
	fmt.Println()
	fmt.Println("Chart build options:")
	fmt.Println("  --chart-dir <path>      Chart directory path (required)")
	fmt.Println("  --push                  Push chart to registry")
	fmt.Println()
	fmt.Println("Common options:")
	fmt.Println("  --project-root <path>   Override project root")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  IMAGE_NAME                      Image name (default: astro-data/<module>, '/' -> '-')")
	fmt.Println("  CONTAINER_REGISTRY              Registry URL (default: auto-detect k3s or localhost:5000)")
	fmt.Println("  CONTAINER_REGISTRY_USERNAME     Registry username")
	fmt.Println("  CONTAINER_REGISTRY_PASSWORD     Registry password")
	fmt.Println("  REGISTRY_PLAIN_HTTP             Use plain HTTP registry (default: false)")
	fmt.Println("  BUILD_ARG_*                     Build arguments (e.g. BUILD_ARG_PIP_INDEX_URL)")
	fmt.Println("  KUBECONFIG                      Kubeconfig path (default: $HOME/.kube/config)")
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
