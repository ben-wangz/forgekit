package version

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

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
