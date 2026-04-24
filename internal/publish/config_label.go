package publish

import (
	"fmt"
	"regexp"
	"strings"
)

var labelKeyPattern = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*(\/[a-z0-9]+([._-][a-z0-9]+)*)*$`)

func (c *ContainerConfig) addLabel(raw string) error {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid --label value %q, expected key=value", raw)
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return fmt.Errorf("invalid --label value %q, key cannot be empty", raw)
	}
	if !labelKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid --label key %q", key)
	}

	value := parts[1]
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		return fmt.Errorf("invalid --label value for key %q, newlines are not allowed", key)
	}

	if _, exists := c.Labels[key]; !exists {
		c.LabelOrder = append(c.LabelOrder, key)
	}
	c.Labels[key] = value

	return nil
}
