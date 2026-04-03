package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	versiontool "github.com/ben-wangz/forgekit/internal/version"
)

func cmdContainerBuild(args []string, projectRoot string) error {
	cfg, err := loadContainerConfig(args, projectRoot)
	if err != nil {
		return err
	}

	containerfilePath := filepath.Join(cfg.ContainerDir, "Containerfile")
	if _, err := os.Stat(containerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Containerfile not found at %s", containerfilePath)
	}

	versionManager, err := versiontool.NewManager(cfg.ProjectRoot)
	if err != nil {
		return err
	}

	publishVersion, err := resolveModulePublishVersion(versionManager, cfg.Module, cfg.Semver)
	if err != nil {
		return fmt.Errorf("failed to resolve publish version: %w", err)
	}

	tags, warnings, err := resolvePublishTags(publishVersion.Value, cfg.Semver, cfg.Push, cfg.MultiTag)
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	primaryTag := publishVersion.Value
	localImage := fmt.Sprintf("%s:%s", cfg.ImageName, primaryTag)
	remoteImages := makeRemoteImages(cfg, tags)

	buildTags := []string{localImage}
	if cfg.Push {
		buildTags = append(buildTags, remoteImages...)
	}

	printBuildInfo(cfg, publishVersion.Mode, primaryTag, localImage, remoteImages, tags)

	fmt.Println()
	fmt.Println("Building image...")
	fmt.Println()

	if err := podmanBuild(cfg, buildTags, containerfilePath); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Build Complete")
	fmt.Println("===========================================")
	fmt.Printf("Image: %s\n", localImage)
	fmt.Println()

	if cfg.Push {
		if err := pushImages(cfg, tags, remoteImages); err != nil {
			return err
		}
	}

	if cfg.LoadK3s {
		k3sRegistry, err := getK3sRegistryAddress()
		if err != nil {
			return fmt.Errorf("failed to get k3s registry address: %w", err)
		}

		k3sRemoteImage := fmt.Sprintf("%s/%s:%s", k3sRegistry, cfg.ImageName, primaryTag)
		if err := loadToK3s(cfg, localImage, k3sRemoteImage); err != nil {
			return fmt.Errorf("failed to load image to k3s: %w", err)
		}
	}

	return nil
}

func printBuildInfo(cfg *ContainerConfig, versionMode, imageTag, localImage string, remoteImages []string, tags []string) {
	fmt.Println("===========================================")
	fmt.Println("Building Container Image")
	fmt.Println("===========================================")
	fmt.Printf("Project root: %s\n", cfg.ProjectRoot)
	fmt.Printf("Container dir: %s\n", cfg.ContainerDir)
	fmt.Printf("Build context: %s\n", cfg.Context)
	fmt.Printf("Version mode: %s\n", versionMode)
	fmt.Printf("Image: %s\n", localImage)
	fmt.Printf("Primary tag: %s\n", imageTag)
	if cfg.Push {
		fmt.Println("Push: enabled")
		fmt.Printf("Push tags: %s\n", strings.Join(tags, ", "))
		for _, image := range remoteImages {
			fmt.Printf("Remote: %s\n", image)
		}
	}

	if len(cfg.BuildArgs) > 0 {
		keys := make([]string, 0, len(cfg.BuildArgs))
		for key := range cfg.BuildArgs {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		fmt.Println()
		for _, key := range keys {
			fmt.Printf("Build arg: %s=%s\n", key, cfg.BuildArgs[key])
		}
	}

	if len(cfg.LabelOrder) > 0 {
		fmt.Println()
		for _, key := range cfg.LabelOrder {
			fmt.Printf("Label: %s=%s\n", key, cfg.Labels[key])
		}
	}

	if len(cfg.BuildArgs) == 0 && len(cfg.LabelOrder) == 0 {
		fmt.Println()
		fmt.Println("No BUILD_ARG_* variables or --label values found")
	}
}

func makeRemoteImages(cfg *ContainerConfig, tags []string) []string {
	images := make([]string, 0, len(tags))
	for _, tag := range tags {
		images = append(images, fmt.Sprintf("%s/%s:%s", cfg.ContainerRegistry, cfg.ImageName, tag))
	}
	return images
}

func pushImages(cfg *ContainerConfig, tags []string, remoteImages []string) error {
	if (cfg.RegistryUsername == "") != (cfg.RegistryPassword == "") {
		return fmt.Errorf("CONTAINER_REGISTRY_USERNAME and CONTAINER_REGISTRY_PASSWORD must be set together")
	}

	fmt.Println("===========================================")
	fmt.Println("Pushing Container Image")
	fmt.Println("===========================================")
	fmt.Println()

	loggedIn := false
	if cfg.RegistryUsername != "" {
		fmt.Println("Logging in to registry...")
		if err := podmanLogin(cfg.ContainerRegistry, cfg.RegistryUsername, cfg.RegistryPassword, cfg.RegistryPlainHTTP); err != nil {
			return fmt.Errorf("failed to login to registry: %w", err)
		}
		fmt.Println()
		loggedIn = true
	}

	pushedTags := make([]string, 0, len(remoteImages))
	for i, remoteImage := range remoteImages {
		fmt.Printf("Pushing image: %s\n", remoteImage)
		if err := podmanPush(remoteImage, cfg.RegistryPlainHTTP); err != nil {
			if loggedIn {
				_ = podmanLogout(cfg.ContainerRegistry)
			}
			if len(pushedTags) > 0 {
				return fmt.Errorf("failed to push image %s: %w (already pushed tags: %s)", remoteImage, err, strings.Join(pushedTags, ", "))
			}
			return fmt.Errorf("failed to push image %s: %w", remoteImage, err)
		}
		pushedTags = append(pushedTags, tags[i])
	}

	if loggedIn {
		fmt.Println()
		fmt.Println("Logging out from registry...")
		if err := podmanLogout(cfg.ContainerRegistry); err != nil {
			return fmt.Errorf("failed to logout from registry: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Push Complete")
	fmt.Println("===========================================")
	fmt.Printf("Pushed tags: %d\n", len(remoteImages))
	for _, image := range remoteImages {
		fmt.Printf("Image: %s\n", image)
	}
	fmt.Println()

	return nil
}

func loadToK3s(cfg *ContainerConfig, localImage, remoteImage string) error {
	fmt.Println("===========================================")
	fmt.Println("Loading to k3s")
	fmt.Println("===========================================")
	fmt.Printf("Local image: %s\n", localImage)
	fmt.Printf("Remote image: %s\n", remoteImage)
	fmt.Println()

	fmt.Println("Tagging image for registry...")
	if err := podmanTag(localImage, remoteImage); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	fmt.Println("Pushing to registry...")
	if err := podmanPush(remoteImage, true); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Println("Pulling from registry to k3s containerd...")
	if err := k3sPullFromRegistry(remoteImage, true); err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Load to k3s Complete")
	fmt.Println("===========================================")
	fmt.Printf("k3s image: %s\n", remoteImage)
	fmt.Println()
	fmt.Println("Use this image in Kubernetes:")
	fmt.Printf("  image: %s\n", remoteImage)
	fmt.Println()

	return nil
}
