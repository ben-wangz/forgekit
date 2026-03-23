package publish

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func helmPackage(chartDir, chartVersion string) (string, error) {
	args := []string{"package", chartDir, "--dependency-update"}
	if chartVersion != "" {
		args = append(args, "--version", chartVersion)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to package chart: %w\n%s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Successfully packaged chart and saved it to: ") {
			return strings.TrimPrefix(line, "Successfully packaged chart and saved it to: "), nil
		}
	}

	return "", fmt.Errorf("failed to parse helm package output")
}

func helmPush(chartPackage, registry string, insecure bool) error {
	ociURL := fmt.Sprintf("oci://%s", registry)
	args := []string{"push", chartPackage, ociURL}
	if insecure {
		args = append(args, "--plain-http")
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func helmRegistryLogin(registry, username, password string, insecure bool) error {
	args := []string{"registry", "login", registry, "--username", username, "--password-stdin"}
	if insecure {
		args = append(args, "--insecure")
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func helmRegistryLogout(registry string) error {
	cmd := exec.Command("helm", "registry", "logout", registry)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
