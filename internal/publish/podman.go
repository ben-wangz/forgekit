package publish

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func podmanBuild(cfg *ContainerConfig, imageTag string, containerfilePath string) error {
	args := []string{
		"build",
		"-t", imageTag,
		"-t", fmt.Sprintf("%s:latest", cfg.ImageName),
		"-f", containerfilePath,
	}

	keys := make([]string, 0, len(cfg.BuildArgs))
	for key := range cfg.BuildArgs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, cfg.BuildArgs[key]))
	}

	args = append(args, cfg.Context)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func podmanTag(sourceTag, targetTag string) error {
	cmd := exec.Command("podman", "tag", sourceTag, targetTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func podmanPush(imageTag string, insecure bool) error {
	args := []string{"push"}
	if insecure {
		args = append(args, "--tls-verify=false")
	}
	args = append(args, imageTag)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func podmanLogin(registry, username, password string, insecure bool) error {
	args := []string{"login", registry, "--username", username, "--password-stdin"}
	if insecure {
		args = append(args, "--tls-verify=false")
	}

	cmd := exec.Command("podman", args...)
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func podmanLogout(registry string) error {
	cmd := exec.Command("podman", "logout", registry)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
