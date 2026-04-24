package version

import "fmt"

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
