package usecases

import (
	"encoding/json"
	"strings"
)

func extractName(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err == nil {
		if v, ok := m["en"].(string); ok {
			return strings.TrimSpace(v)
		}
		if v, ok := m["name"].(string); ok {
			return strings.TrimSpace(v)
		}
		for _, v := range m {
			if vs, ok := v.(string); ok && vs != "" {
				return strings.TrimSpace(vs)
			}
		}
	}
	return ""
}

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
