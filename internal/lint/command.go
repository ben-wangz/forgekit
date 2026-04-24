package lint

import (
	"fmt"
	"path/filepath"

	"github.com/ben-wangz/forgekit/internal/project"
)

func Run(args []string, defaultProjectRoot string) error {
	args, rootOverride, configPath, err := extractFlags(args)
	if err != nil {
		return err
	}

	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		printUsage()
		return nil
	}

	if len(args) > 0 {
		printUsage()
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	projectRoot := defaultProjectRoot
	if rootOverride != "" {
		projectRoot = rootOverride
	}

	resolvedRoot, err := project.ResolveRoot(projectRoot)
	if err != nil {
		return err
	}

	if configPath == "" {
		configPath = filepath.Join(resolvedRoot, "lint.yaml")
	} else if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(resolvedRoot, configPath)
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	config, err := loadConfig(absConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	baseDir := filepath.Dir(absConfigPath)

	if len(config.Commands) > 0 {
		if err := runCommands(baseDir, config.Commands); err != nil {
			return err
		}
	}

	if len(config.MaxLinesByExt) > 0 {
		if err := checkFiles(baseDir, config); err != nil {
			return err
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: forgekit lint [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config <path>         Lint config path (default: <project-root>/lint.yaml)")
	fmt.Println("  --project-root <path>   Override project root")
	fmt.Println("  --help, -h              Print help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  forgekit lint")
	fmt.Println("  forgekit lint --config lint.yaml")
}

func extractFlags(args []string) ([]string, string, string, error) {
	filtered := make([]string, 0, len(args))
	projectRoot := ""
	configPath := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project-root":
			if i+1 >= len(args) {
				return nil, "", "", fmt.Errorf("--project-root requires a value")
			}
			projectRoot = args[i+1]
			i++
		case "--config":
			if i+1 >= len(args) {
				return nil, "", "", fmt.Errorf("--config requires a value")
			}
			configPath = args[i+1]
			i++
		default:
			filtered = append(filtered, args[i])
		}
	}

	return filtered, projectRoot, configPath, nil
}
