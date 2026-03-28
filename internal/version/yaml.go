package version

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func readYAMLValue(filePath, keyPath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var content map[string]any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return "", err
	}

	keys := strings.Split(keyPath, ".")
	current := content

	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				return fmt.Sprintf("%v", val), nil
			}
			return "", fmt.Errorf("key not found: %s", keyPath)
		}

		next, ok := current[key].(map[string]any)
		if !ok {
			return "", fmt.Errorf("key not found: %s", keyPath)
		}
		current = next
	}

	return "", fmt.Errorf("key not found: %s", keyPath)
}

func updateYAMLValue(filePath, keyPath, value string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	patched, err := patchYAMLScalarValue(data, keyPath, value)
	if err != nil {
		return fmt.Errorf("file=%s keyPath=%s: %w", filePath, keyPath, err)
	}

	if bytes.Equal(data, patched) {
		return nil
	}

	return os.WriteFile(filePath, patched, 0644)
}
