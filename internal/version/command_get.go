package version

import "fmt"

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
