package version

func patchYAMLScalarValue(content []byte, keyPath, newValue string) ([]byte, error) {
	keys, err := splitKeyPath(keyPath)
	if err != nil {
		return nil, err
	}

	targetNode, err := findUniqueKeyPathNode(content, keys)
	if err != nil {
		return nil, err
	}

	start, end, err := locateScalarTokenRange(content, targetNode)
	if err != nil {
		return nil, err
	}

	replacement := renderScalarLiteral(targetNode.Style, newValue)
	patched := make([]byte, 0, len(content)-(end-start)+len(replacement))
	patched = append(patched, content[:start]...)
	patched = append(patched, replacement...)
	patched = append(patched, content[end:]...)

	return patched, nil
}
