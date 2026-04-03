package publish

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ben-wangz/forgekit/internal/project"
)

var labelKeyPattern = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*(\/[a-z0-9]+([._-][a-z0-9]+)*)*$`)

type ContainerConfig struct {
	ContainerDir      string
	Module            string
	Push              bool
	Semver            bool
	MultiTag          bool
	LoadK3s           bool
	Context           string
	ImageName         string
	ContainerRegistry string
	RegistryUsername  string
	RegistryPassword  string
	RegistryPlainHTTP bool
	BuildArgs         map[string]string
	Labels            map[string]string
	LabelOrder        []string
	ProjectRoot       string
}

func loadContainerConfig(args []string, projectRoot string) (*ContainerConfig, error) {
	registryExplicitlySet := strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY")) != ""
	plainHTTPExplicitlySet := strings.TrimSpace(os.Getenv("REGISTRY_PLAIN_HTTP")) != ""
	imageNameExplicitlySet := strings.TrimSpace(os.Getenv("IMAGE_NAME")) != ""

	cfg := &ContainerConfig{
		ImageName:         strings.TrimSpace(os.Getenv("IMAGE_NAME")),
		ContainerRegistry: getContainerRegistry(),
		RegistryUsername:  strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY_USERNAME")),
		RegistryPassword:  strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY_PASSWORD")),
		BuildArgs:         make(map[string]string),
		Labels:            make(map[string]string),
		LabelOrder:        []string{},
		ProjectRoot:       projectRoot,
	}

	if !registryExplicitlySet && !plainHTTPExplicitlySet {
		cfg.RegistryPlainHTTP = true
	} else {
		cfg.RegistryPlainHTTP = strings.TrimSpace(os.Getenv("REGISTRY_PLAIN_HTTP")) == "true"
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--container-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--container-dir requires a value")
			}
			cfg.ContainerDir = args[i+1]
			i++
		case "--module":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--module requires a value")
			}
			cfg.Module = args[i+1]
			i++
		case "--push":
			cfg.Push = true
		case "--semver":
			cfg.Semver = true
		case "--multi-tag":
			cfg.MultiTag = true
		case "--load-k3s":
			cfg.LoadK3s = true
		case "--context":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--context requires a value")
			}
			cfg.Context = args[i+1]
			i++
		case "--label":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--label requires a value")
			}
			if err := cfg.addLabel(args[i+1]); err != nil {
				return nil, err
			}
			i++
		default:
			return nil, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	if cfg.MultiTag && !cfg.Semver {
		return nil, fmt.Errorf("--multi-tag requires --semver")
	}

	if cfg.MultiTag && !cfg.Push {
		return nil, fmt.Errorf("--multi-tag requires --push")
	}

	if cfg.ContainerDir == "" {
		return nil, fmt.Errorf("--container-dir is required")
	}
	if cfg.Module == "" {
		return nil, fmt.Errorf("--module is required")
	}

	if !imageNameExplicitlySet {
		cfg.ImageName = defaultImageNameFromModule(cfg.Module)
	}
	if cfg.ImageName == "" {
		return nil, fmt.Errorf("image name resolved from module is empty")
	}

	cfg.ContainerDir = project.ResolvePath(projectRoot, cfg.ContainerDir)
	if cfg.Context == "" {
		cfg.Context = projectRoot
	} else {
		cfg.Context = project.ResolvePath(projectRoot, cfg.Context)
	}

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "BUILD_ARG_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				argName := strings.TrimPrefix(parts[0], "BUILD_ARG_")
				cfg.BuildArgs[argName] = parts[1]
			}
		}
	}

	return cfg, nil
}

func (c *ContainerConfig) addLabel(raw string) error {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid --label value %q, expected key=value", raw)
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return fmt.Errorf("invalid --label value %q, key cannot be empty", raw)
	}
	if !labelKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid --label key %q", key)
	}

	value := parts[1]
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		return fmt.Errorf("invalid --label value for key %q, newlines are not allowed", key)
	}

	if _, exists := c.Labels[key]; !exists {
		c.LabelOrder = append(c.LabelOrder, key)
	}
	c.Labels[key] = value

	return nil
}

type ChartConfig struct {
	ChartDir                     string
	Push                         bool
	Semver                       bool
	ChartRegistry                string
	ChartRegistryHost            string
	RegistryPlainHTTP            bool
	Username                     string
	Password                     string
	ProjectRoot                  string
	RegistryFallbackDeprecated   bool
	CredentialFallbackDeprecated bool
}

func loadChartConfig(args []string, projectRoot string) (*ChartConfig, error) {
	registryExplicitlySet := strings.TrimSpace(os.Getenv("CONTAINER_REGISTRY")) != "" || strings.TrimSpace(os.Getenv("CHART_REGISTRY")) != ""
	plainHTTPExplicitlySet := strings.TrimSpace(os.Getenv("REGISTRY_PLAIN_HTTP")) != ""

	chartRegistry, registryFallbackDeprecated, err := resolveChartRegistry()
	if err != nil {
		return nil, err
	}

	username, password, credentialFallbackDeprecated, err := resolveChartCredentials()
	if err != nil {
		return nil, err
	}

	cfg := &ChartConfig{
		ChartRegistry:                chartRegistry,
		ChartRegistryHost:            registryHost(chartRegistry),
		Username:                     username,
		Password:                     password,
		ProjectRoot:                  projectRoot,
		RegistryFallbackDeprecated:   registryFallbackDeprecated,
		CredentialFallbackDeprecated: credentialFallbackDeprecated,
	}

	if !registryExplicitlySet && !plainHTTPExplicitlySet {
		cfg.RegistryPlainHTTP = true
	} else {
		cfg.RegistryPlainHTTP = strings.TrimSpace(os.Getenv("REGISTRY_PLAIN_HTTP")) == "true"
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--chart-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--chart-dir requires a value")
			}
			cfg.ChartDir = args[i+1]
			i++
		case "--push":
			cfg.Push = true
		case "--semver":
			cfg.Semver = true
		default:
			return nil, fmt.Errorf("unknown option: %s", args[i])
		}

	}

	if cfg.ChartDir == "" {
		return nil, fmt.Errorf("--chart-dir is required")
	}

	cfg.ChartDir = project.ResolvePath(projectRoot, cfg.ChartDir)
	return cfg, nil
}

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
