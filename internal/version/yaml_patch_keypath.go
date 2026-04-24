package version

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func splitKeyPath(keyPath string) ([]string, error) {
	parts := strings.Split(keyPath, ".")
	keys := make([]string, 0, len(parts))

	for _, key := range parts {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid keyPath: %s", keyPath)
		}
		keys = append(keys, trimmed)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("invalid keyPath: %s", keyPath)
	}

	return keys, nil
}

func findUniqueKeyPathNode(content []byte, keys []string) (*yaml.Node, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	matches := collectKeyPathNodes(&root, keys)
	if len(matches) == 0 {
		return nil, fmt.Errorf("keyPath not found")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("keyPath matched multiple nodes")
	}

	targetNode := matches[0]
	if targetNode.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("target is not scalar")
	}

	return targetNode, nil
}

func collectKeyPathNodes(node *yaml.Node, keys []string) []*yaml.Node {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.DocumentNode {
		matches := []*yaml.Node{}
		for _, child := range node.Content {
			matches = append(matches, collectKeyPathNodes(child, keys)...)
		}
		return matches
	}

	if len(keys) == 0 {
		return []*yaml.Node{node}
	}

	if node.Kind != yaml.MappingNode {
		return nil
	}

	matches := []*yaml.Node{}
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		if keyNode.Kind != yaml.ScalarNode {
			continue
		}

		if keyNode.Value != keys[0] {
			continue
		}

		matches = append(matches, collectKeyPathNodes(valueNode, keys[1:])...)
	}

	return matches
}
