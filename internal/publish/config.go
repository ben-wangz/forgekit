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
	LoadK3s           bool
	Context           string
	ImageName         string
	ContainerRegistry string
	RegistryUsername  string
	RegistryPassword  string
	RegistryPlainHTTP bool
	BuildArgs         map[string]string
	ProjectRoot       string
}

func loadContainerConfig(args []string, projectRoot string) (*ContainerConfig, error) {
	registryExplicitlySet := os.Getenv("CONTAINER_REGISTRY") != ""
	plainHTTPExplicitlySet := os.Getenv("REGISTRY_PLAIN_HTTP") != ""
	imageNameExplicitlySet := os.Getenv("IMAGE_NAME") != ""

	cfg := &ContainerConfig{
		ImageName:         os.Getenv("IMAGE_NAME"),
		ContainerRegistry: getContainerRegistry(),
		RegistryUsername:  os.Getenv("CONTAINER_REGISTRY_USERNAME"),
		RegistryPassword:  os.Getenv("CONTAINER_REGISTRY_PASSWORD"),
		BuildArgs:         make(map[string]string),
		ProjectRoot:       projectRoot,
	}

	if !registryExplicitlySet && !plainHTTPExplicitlySet {
		cfg.RegistryPlainHTTP = true
	} else {
		cfg.RegistryPlainHTTP = os.Getenv("REGISTRY_PLAIN_HTTP") == "true"
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
		case "--load-k3s":
			cfg.LoadK3s = true
		case "--context":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--context requires a value")
			}
			cfg.Context = args[i+1]
			i++
		default:
			return nil, fmt.Errorf("unknown option: %s", args[i])
		}
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

type ChartConfig struct {
	ChartDir          string
	Push              bool
	Semver            bool
	ContainerRegistry string
	RegistryPlainHTTP bool
	Username          string
	Password          string
	ProjectRoot       string
}

func loadChartConfig(args []string, projectRoot string) (*ChartConfig, error) {
	registryExplicitlySet := os.Getenv("CONTAINER_REGISTRY") != ""
	plainHTTPExplicitlySet := os.Getenv("REGISTRY_PLAIN_HTTP") != ""

	cfg := &ChartConfig{
		ContainerRegistry: getContainerRegistry(),
		Username:          os.Getenv("CONTAINER_REGISTRY_USERNAME"),
		Password:          os.Getenv("CONTAINER_REGISTRY_PASSWORD"),
		ProjectRoot:       projectRoot,
	}

	if !registryExplicitlySet && !plainHTTPExplicitlySet {
		cfg.RegistryPlainHTTP = true
	} else {
		cfg.RegistryPlainHTTP = os.Getenv("REGISTRY_PLAIN_HTTP") == "true"
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

func getContainerRegistry() string {
	if registry := os.Getenv("CONTAINER_REGISTRY"); registry != "" {
		return registry
	}

	registry, err := getK3sRegistryAddress()
	if err == nil {
		return registry
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
