package publish

import (
	"fmt"
	"os"
	"strings"
)

func resolveChartRegistry() (string, bool, error) {
	chartRegistry := strings.TrimSpace(os.Getenv("CHART_REGISTRY"))
	if chartRegistry != "" {
		if strings.HasPrefix(chartRegistry, "oci://") {
			return "", false, fmt.Errorf("CHART_REGISTRY must not include oci:// prefix")
		}

		normalized := normalizeRegistry(chartRegistry)
		if normalized == "" {
			return "", false, fmt.Errorf("CHART_REGISTRY is invalid")
		}

		return normalized, false, nil
	}

	containerRegistryEnv := strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY"))
	if containerRegistryEnv != "" {
		containerRegistry := normalizeRegistry(containerRegistryEnv)
		return appendChartsSuffix(containerRegistry), true, nil
	}

	autoDetected := normalizeRegistry(getContainerRegistry())
	return appendChartsSuffix(autoDetected), true, nil
}

func resolveChartCredentials() (string, string, bool, error) {
	chartUser := strings.TrimSpace(os.Getenv("CHART_REGISTRY_USERNAME"))
	chartPass := strings.TrimSpace(os.Getenv("CHART_REGISTRY_PASSWORD"))

	if (chartUser == "") != (chartPass == "") {
		return "", "", false, fmt.Errorf("CHART_REGISTRY_USERNAME and CHART_REGISTRY_PASSWORD must be set together")
	}

	if chartUser != "" {
		return chartUser, chartPass, false, nil
	}

	containerUser := strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY_USERNAME"))
	containerPass := strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY_PASSWORD"))

	if containerUser == "" && containerPass == "" {
		return "", "", false, nil
	}

	if (containerUser == "") != (containerPass == "") {
		return "", "", false, fmt.Errorf("CONTAINER_REGISTRY_USERNAME and CONTAINER_REGISTRY_PASSWORD must be set together when used as chart credential fallback")
	}

	return containerUser, containerPass, true, nil
}

func appendChartsSuffix(registry string) string {
	trimmed := normalizeRegistry(registry)
	if strings.HasSuffix(trimmed, "/charts") {
		return trimmed
	}
	return trimmed + "/charts"
}

func normalizeRegistry(registry string) string {
	trimmed := strings.TrimSpace(registry)
	for strings.HasSuffix(trimmed, "/") {
		trimmed = strings.TrimSuffix(trimmed, "/")
	}
	return trimmed
}

func registryHost(registry string) string {
	parts := strings.SplitN(registry, "/", 2)
	return parts[0]
}
