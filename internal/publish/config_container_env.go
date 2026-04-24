package publish

import (
	"os"
	"strings"
)

func getContainerRegistry() string {
	if registry := normalizeRegistry(strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY"))); registry != "" {
		return registry
	}

	registry, err := getK3sRegistryAddress()
	if err == nil {
		return normalizeRegistry(registry)
	}

	return "localhost:5000"
}

func defaultImageNameFromModule(module string) string {
	trimmed := strings.Trim(module, "/")
	if trimmed == "" {
		return ""
	}

	normalized := strings.ReplaceAll(trimmed, "/", "-")
	return "astro-data/" + normalized
}

func collectBuildArgs(target map[string]string) {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "BUILD_ARG_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				argName := strings.TrimPrefix(parts[0], "BUILD_ARG_")
				target[argName] = parts[1]
			}
		}
	}
}
