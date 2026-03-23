package version

import (
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

	var content map[string]any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return err
	}

	keys := strings.Split(keyPath, ".")
	current := content

	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			break
		}

		next, ok := current[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[key] = next
		}
		current = next
	}

	output, err := yaml.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, output, 0644)
}
