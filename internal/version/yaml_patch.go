package version

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func patchYAMLScalarValue(content []byte, keyPath, newValue string) ([]byte, error) {
	keys, err := splitKeyPath(keyPath)
	if err != nil {
		return nil, err
	}

	targetNode, err := findUniqueKeyPathNode(content, keys)
	if err != nil {
		return nil, err
	}

	if targetNode.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("target is not scalar")
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

	return matches[0], nil
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

func renderScalarLiteral(style yaml.Style, value string) string {
	if style&yaml.DoubleQuotedStyle != 0 {
		return strconv.Quote(value)
	}

	if style&yaml.SingleQuotedStyle != 0 {
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	}

	if isSafePlainScalar(value) {
		return value
	}

	return strconv.Quote(value)
}

func isSafePlainScalar(value string) bool {
	if strings.TrimSpace(value) != value || value == "" {
		return false
	}

	lower := strings.ToLower(value)
	reserved := map[string]struct{}{
		"null":  {},
		"~":     {},
		"true":  {},
		"false": {},
		"yes":   {},
		"no":    {},
		"on":    {},
		"off":   {},
	}
	if _, ok := reserved[lower]; ok {
		return false
	}

	if strings.ContainsAny(value, "\r\n\t") {
		return false
	}
	if strings.Contains(value, ": ") || strings.Contains(value, " #") {
		return false
	}
	if strings.HasPrefix(value, "-") || strings.HasPrefix(value, "?") || strings.HasPrefix(value, ":") {
		return false
	}

	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}

		switch r {
		case '.', '_', '-', '/':
			continue
		default:
			return false
		}
	}

	return true
}

func locateScalarTokenRange(content []byte, node *yaml.Node) (int, int, error) {
	if node.Line <= 0 || node.Column <= 0 {
		return 0, 0, fmt.Errorf("target scalar has no valid line/column info")
	}

	start, err := lineColumnToOffset(content, node.Line, node.Column)
	if err != nil {
		return 0, 0, err
	}

	if start >= len(content) {
		return 0, 0, fmt.Errorf("target position is out of range")
	}

	style := node.Style
	if style&yaml.DoubleQuotedStyle != 0 {
		end, err := scanDoubleQuotedEnd(content, start)
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	if style&yaml.SingleQuotedStyle != 0 {
		end, err := scanSingleQuotedEnd(content, start)
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	if style&yaml.LiteralStyle != 0 || style&yaml.FoldedStyle != 0 {
		return 0, 0, fmt.Errorf("block scalar is not supported for keyPath patch")
	}

	end, err := scanPlainScalarEnd(content, start)
	if err != nil {
		return 0, 0, err
	}

	return start, end, nil
}

func lineColumnToOffset(content []byte, line, column int) (int, error) {
	if line < 1 || column < 1 {
		return 0, fmt.Errorf("invalid line/column: %d:%d", line, column)
	}

	lineStart := 0
	currentLine := 1
	for currentLine < line {
		next := bytesIndexByte(content, '\n', lineStart)
		if next < 0 {
			return 0, fmt.Errorf("line out of range: %d", line)
		}
		lineStart = next + 1
		currentLine++
	}

	lineEnd := bytesIndexByte(content, '\n', lineStart)
	if lineEnd < 0 {
		lineEnd = len(content)
	}

	offset := lineStart
	runesToAdvance := column - 1
	for i := 0; i < runesToAdvance; i++ {
		if offset >= lineEnd {
			return 0, fmt.Errorf("column out of range: %d:%d", line, column)
		}
		_, size := utf8.DecodeRune(content[offset:lineEnd])
		if size <= 0 {
			return 0, fmt.Errorf("invalid utf-8 sequence at %d:%d", line, column)
		}
		offset += size
	}

	return offset, nil
}

func bytesIndexByte(content []byte, target byte, start int) int {
	for i := start; i < len(content); i++ {
		if content[i] == target {
			return i
		}
	}
	return -1
}

func scanDoubleQuotedEnd(content []byte, start int) (int, error) {
	if content[start] != '"' {
		return 0, fmt.Errorf("double-quoted scalar token not found at expected position")
	}

	i := start + 1
	for i < len(content) {
		switch content[i] {
		case '\\':
			i += 2
		case '"':
			return i + 1, nil
		case '\n', '\r':
			return 0, fmt.Errorf("unterminated double-quoted scalar")
		default:
			i++
		}
	}

	return 0, fmt.Errorf("unterminated double-quoted scalar")
}

func scanSingleQuotedEnd(content []byte, start int) (int, error) {
	if content[start] != '\'' {
		return 0, fmt.Errorf("single-quoted scalar token not found at expected position")
	}

	i := start + 1
	for i < len(content) {
		switch content[i] {
		case '\'':
			if i+1 < len(content) && content[i+1] == '\'' {
				i += 2
				continue
			}
			return i + 1, nil
		case '\n', '\r':
			return 0, fmt.Errorf("unterminated single-quoted scalar")
		default:
			i++
		}
	}

	return 0, fmt.Errorf("unterminated single-quoted scalar")
}

func scanPlainScalarEnd(content []byte, start int) (int, error) {
	lineEnd := bytesIndexByte(content, '\n', start)
	if lineEnd < 0 {
		lineEnd = len(content)
	}

	segment := content[start:lineEnd]
	commentStart := findCommentStart(segment)
	if commentStart >= 0 {
		segment = segment[:commentStart]
	}

	trimmedLen := len(segment)
	for trimmedLen > 0 {
		c := segment[trimmedLen-1]
		if c == ' ' || c == '\t' || c == '\r' {
			trimmedLen--
			continue
		}
		break
	}

	if trimmedLen <= 0 {
		return 0, fmt.Errorf("plain scalar token not found at expected position")
	}

	return start + trimmedLen, nil
}

func findCommentStart(segment []byte) int {
	for i := 0; i < len(segment); i++ {
		if segment[i] != '#' {
			continue
		}
		if i == 0 || segment[i-1] == ' ' || segment[i-1] == '\t' {
			return i
		}
	}

	return -1
}
