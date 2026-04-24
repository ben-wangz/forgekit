package main

import (
	"fmt"
	"os"
	"strings"

	lintcmd "github.com/ben-wangz/forgekit/internal/lint"
	"github.com/ben-wangz/forgekit/internal/publish"
	versioncmd "github.com/ben-wangz/forgekit/internal/version"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	projectRoot, rest, showVersion, showHelp, err := parseGlobalFlags(args)
	if err != nil {
		return err
	}

	if showVersion {
		fmt.Printf("forgekit %s (%s)\n", version, commit)
		return nil
	}

	if showHelp || len(rest) == 0 {
		printUsage()
		return nil
	}

	switch rest[0] {
	case "lint":
		return lintcmd.Run(rest[1:], projectRoot)
	case "version":
		return versioncmd.Run(rest[1:], projectRoot)
	case "publish":
		return publish.Run(rest[1:], projectRoot)
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", rest[0])
	}
}

func parseGlobalFlags(args []string) (string, []string, bool, bool, error) {
	projectRoot := ""
	showVersion := false
	showHelp := false

	for len(args) > 0 {
		arg := args[0]
		if !strings.HasPrefix(arg, "-") {
			break
		}

		switch arg {
		case "--project-root":
			if len(args) < 2 {
				return "", nil, false, false, fmt.Errorf("--project-root requires a value")
			}
			projectRoot = args[1]
			args = args[2:]
		case "--version", "-v":
			showVersion = true
			args = args[1:]
		case "--help", "-h":
			showHelp = true
			args = args[1:]
		default:
			return "", nil, false, false, fmt.Errorf("unknown global option: %s", arg)
		}
	}

	return projectRoot, args, showVersion, showHelp, nil
}

func printUsage() {
	fmt.Println("Usage: forgekit [global options] <command> [args]")
	fmt.Println()
	fmt.Println("Global options:")
	fmt.Println("  --project-root <path>   Project root directory")
	fmt.Println("  --version, -v           Print forgekit version")
	fmt.Println("  --help, -h              Print help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  lint                    Code style checks")
	fmt.Println("  version                 Version management")
	fmt.Println("  publish                 Container/chart publishing")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  forgekit lint")
	fmt.Println("  forgekit lint --config lint.yaml")
	fmt.Println("  forgekit version get")
	fmt.Println("  forgekit version get catalog/ingest --git")
	fmt.Println("  forgekit publish container build --container-dir catalog/ingest/container --module catalog/ingest")
	fmt.Println("  forgekit publish chart build --chart-dir operator/chart --push")
}
