package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

	imageTag, err := versionManager.ModuleGitVersion(cfg.Module)
	if err != nil {
		return fmt.Errorf("failed to get git-version: %w", err)
	}

	localImage := fmt.Sprintf("%s:%s", cfg.ImageName, imageTag)
	remoteImage := fmt.Sprintf("%s/%s:%s", cfg.ContainerRegistry, cfg.ImageName, imageTag)

	printBuildInfo(cfg, imageTag, localImage, remoteImage)

	fmt.Println()
	fmt.Println("Building image...")
	fmt.Println()

	if err := podmanBuild(cfg, localImage, containerfilePath); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Build Complete")
	fmt.Println("===========================================")
	fmt.Printf("Image: %s\n", localImage)
	fmt.Printf("Also tagged as: %s:latest\n", cfg.ImageName)
	fmt.Println()

	if cfg.Push {
		if err := pushImage(cfg, localImage, remoteImage); err != nil {
			return err
		}
	}

	if cfg.LoadK3s {
		k3sRegistry, err := getK3sRegistryAddress()
		if err != nil {
			return fmt.Errorf("failed to get k3s registry address: %w", err)
		}

		k3sRemoteImage := fmt.Sprintf("%s/%s:%s", k3sRegistry, cfg.ImageName, imageTag)
		if err := loadToK3s(cfg, localImage, k3sRemoteImage); err != nil {
			return fmt.Errorf("failed to load image to k3s: %w", err)
		}
	}

	return nil
}

func printBuildInfo(cfg *ContainerConfig, imageTag, localImage, remoteImage string) {
	fmt.Println("===========================================")
	fmt.Println("Building Container Image")
	fmt.Println("===========================================")
	fmt.Printf("Project root: %s\n", cfg.ProjectRoot)
	fmt.Printf("Container dir: %s\n", cfg.ContainerDir)
	fmt.Printf("Build context: %s\n", cfg.Context)
	fmt.Printf("Image: %s\n", localImage)
	fmt.Printf("Tag: %s\n", imageTag)
	if cfg.Push {
		fmt.Println("Push: enabled")
		fmt.Printf("Remote: %s\n", remoteImage)
	}

	if len(cfg.BuildArgs) == 0 {
		fmt.Println()
		fmt.Println("No BUILD_ARG_* variables found")
		return
	}

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

func pushImage(cfg *ContainerConfig, localImage, remoteImage string) error {
	fmt.Println("===========================================")
	fmt.Println("Pushing Container Image")
	fmt.Println("===========================================")
	fmt.Printf("Remote image: %s\n", remoteImage)
	fmt.Println()

	if cfg.RegistryUsername != "" && cfg.RegistryPassword != "" {
		fmt.Println("Logging in to registry...")
		if err := podmanLogin(cfg.ContainerRegistry, cfg.RegistryUsername, cfg.RegistryPassword, cfg.RegistryPlainHTTP); err != nil {
			return fmt.Errorf("failed to login to registry: %w", err)
		}
		fmt.Println()
	}

	fmt.Println("Tagging image...")
	if err := podmanTag(localImage, remoteImage); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	fmt.Println("Pushing image...")
	if err := podmanPush(remoteImage, cfg.RegistryPlainHTTP); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	if cfg.RegistryUsername != "" && cfg.RegistryPassword != "" {
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
	fmt.Printf("Image: %s\n", remoteImage)
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
