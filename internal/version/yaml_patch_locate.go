package version

import (
	"fmt"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

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
