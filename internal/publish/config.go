package publish

import (
	"fmt"
	"os"
	"strings"

	"github.com/ben-wangz/forgekit/internal/project"
)

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

	collectBuildArgs(cfg.BuildArgs)

	return cfg, nil
}

type ChartConfig struct {
	ChartDir                     string
	Push                         bool
	Semver                       bool
	MultiTag                     bool
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
		case "--multi-tag":
			cfg.MultiTag = true
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

	if cfg.ChartDir == "" {
		return nil, fmt.Errorf("--chart-dir is required")
	}

	cfg.ChartDir = project.ResolvePath(projectRoot, cfg.ChartDir)
	return cfg, nil
}
