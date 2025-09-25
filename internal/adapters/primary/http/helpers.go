package http

import (
	"strings"
)

func normalizePlayerTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tag
	}
	if strings.HasPrefix(tag, "%23") {
		return tag
	}
	if strings.HasPrefix(tag, "#") {
		return "%23" + strings.TrimPrefix(tag, "#")
	}
	return "%23" + tag
}
