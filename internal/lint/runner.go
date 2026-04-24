package lint

import (
	"fmt"
	"os"
	"os/exec"
)

func runCommands(baseDir string, commands []Command) error {
	for _, lintCommand := range commands {
		fmt.Printf("Running: %s\n", lintCommand.Name)

		command := exec.Command(lintCommand.Cmd, lintCommand.Args...)
		command.Dir = baseDir
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			return fmt.Errorf("%s failed: %w", lintCommand.Name, err)
		}

		fmt.Printf("✓ %s passed\n\n", lintCommand.Name)
	}

	return nil
}
