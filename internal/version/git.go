package version

import (
	"fmt"
	"os/exec"
	"strings"
)

func buildGitVersion(version, commitID string, dirty bool) string {
	dirtySuffix := ""
	if dirty {
		dirtySuffix = "-dirty"
	}

	return fmt.Sprintf("%s-%s%s", version, commitID, dirtySuffix)
}

func getGitCommitID(projectRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short=7", "HEAD")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit ID: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func isGitDirty(projectRoot string) (bool, error) {
	refreshCmd := exec.Command("git", "update-index", "-q", "--refresh")
	refreshCmd.Dir = projectRoot
	if err := refreshCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to refresh git index: %w", err)
	}

	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	cmd.Dir = projectRoot
	err := cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			return true, nil
		}
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = projectRoot
	untrackedOutput, err := untrackedCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check untracked files: %w", err)
	}

	if strings.TrimSpace(string(untrackedOutput)) != "" {
		return true, nil
	}

	return false, nil
}
