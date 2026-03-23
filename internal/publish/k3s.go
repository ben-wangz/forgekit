package publish

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func checkK3sAvailable() error {
	if _, err := exec.LookPath("k3s"); err != nil {
		return fmt.Errorf("k3s command not found")
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfig = fmt.Sprintf("%s/.kube/config", homeDir)
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("KUBECONFIG not found: %s", kubeconfig)
	}

	return nil
}

func k3sPullFromRegistry(imageTag string, insecure bool) error {
	if err := checkK3sAvailable(); err != nil {
		return err
	}

	args := []string{"ctr", "images", "pull", "--snapshotter", "native", "--all-platforms"}
	if insecure {
		args = append(args, "--plain-http")
	}
	args = append(args, imageTag)

	cmd := exec.Command("k3s", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getK3sRegistryAddress() (string, error) {
	cmd := exec.Command("kubectl", "get", "nodes",
		"--selector=node-role.kubernetes.io/control-plane",
		"-o", "jsonpath={$.items[0].status.addresses[?(@.type==\"InternalIP\")].address}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get master node IP: %w", err)
	}

	host := strings.TrimSpace(string(output))
	if host == "" {
		return "", fmt.Errorf("master node IP not found")
	}

	cmd = exec.Command("kubectl", "-n", "basic-components", "get", "service",
		"container-registry-docker-registry",
		"-o", "jsonpath={.spec.ports[0].nodePort}")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get registry service port: %w", err)
	}

	port := strings.TrimSpace(string(output))
	if port == "" {
		return "", fmt.Errorf("registry service port not found")
	}

	return fmt.Sprintf("%s:%s", host, port), nil
}
