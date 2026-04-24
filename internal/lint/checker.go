package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func checkFiles(baseDir string, config *Config) error {
	failed := false

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		if shouldExclude(relPath, config.Exclude) {
			return nil
		}

		if !shouldInclude(relPath, config.Include) {
			return nil
		}

		maxLines, ok := getMaxLines(path, config)
		if !ok {
			return nil
		}

		lineCount, err := countLines(path)
		if err != nil {
			return err
		}

		if lineCount > maxLines {
			fmt.Printf("%s: %d lines (exceeds %d)\n", relPath, lineCount, maxLines)
			failed = true
		}

		return nil
	})
	if err != nil {
		return err
	}

	if failed {
		return fmt.Errorf("some files exceed line limits")
	}

	return nil
}

func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := matchPattern(pattern, path)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func shouldInclude(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	for _, pattern := range patterns {
		matched, err := matchPattern(pattern, filepath.Base(path))
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func getMaxLines(path string, config *Config) (int, bool) {
	ext := filepath.Ext(path)
	maxLines, ok := config.MaxLinesByExt[ext]
	if !ok {
		return 0, false
	}

	return maxLines, true
}

func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	lineCount := strings.Count(string(data), "\n")
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		lineCount++
	}

	return lineCount, nil
}

func matchPattern(pattern, value string) (bool, error) {
	patternRegex := globToRegex(filepath.ToSlash(pattern))
	compiled, err := regexp.Compile(patternRegex)
	if err != nil {
		return false, err
	}

	return compiled.MatchString(filepath.ToSlash(value)), nil
}

func globToRegex(pattern string) string {
	var builder strings.Builder
	builder.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		if char == '*' {
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				builder.WriteString(".*")
				i++
				continue
			}
			builder.WriteString("[^/]*")
			continue
		}

		switch char {
		case '?':
			builder.WriteString(".")
		case '.', '+', '(', ')', '[', ']', '{', '}', '^', '$', '|', '\\':
			builder.WriteString("\\")
			builder.WriteByte(char)
		default:
			builder.WriteByte(char)
		}
	}

	builder.WriteString("$")
	return builder.String()
}
