package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dauletkhan/coc/internal/catalog"
	"github.com/dauletkhan/coc/internal/coc"
)

type HeroEquipmentsHandler struct {
	coc     *coc.Client
	catalog catalog.EquipmentCatalog
}

func NewHeroEquipmentsHandler(cocClient *coc.Client, cat catalog.EquipmentCatalog) *HeroEquipmentsHandler {
	return &HeroEquipmentsHandler{coc: cocClient, catalog: cat}
}

func (h *HeroEquipmentsHandler) Register(r *gin.Engine) {
	r.GET("/v1/players/:tag/hero-equipments", h.getHeroEquipments)
}

func normalizePlayerTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tag
	}
	// Accept raw '#', already-encoded '%23', or bare without prefix
	if strings.HasPrefix(tag, "%23") {
		return tag
	}
	if strings.HasPrefix(tag, "#") {
		return "%23" + strings.TrimPrefix(tag, "#")
	}
	// No prefix provided; prepend encoded '#'
	return "%23" + tag
}

func extractName(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err == nil {
		// Try common keys
		if v, ok := m["name"].(string); ok {
			return strings.TrimSpace(v)
		}
		if v, ok := m["en"].(string); ok {
			return strings.TrimSpace(v)
		}
		// Fallback: first string value
		for _, v := range m {
			if vs, ok := v.(string); ok && vs != "" {
				return strings.TrimSpace(vs)
			}
		}
	}
	return ""
}

func (h *HeroEquipmentsHandler) getHeroEquipments(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tag"})
		return
	}
	nTag := normalizePlayerTag(tag)
	// Basic sanity: ensure path-safe (already encoded '#')
	if _, err := url.PathUnescape(nTag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
	defer cancel()

	body, status, err := h.coc.GetPlayerRaw(ctx, nTag)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if status >= 400 {
		c.Data(status, "application/json", body)
		return
	}

	// parse minimal fields using std lib to avoid strict schema coupling
	type equipment struct {
		Name     json.RawMessage `json:"name"`
		Level    int             `json:"level"`
		MaxLevel int             `json:"maxLevel"`
	}
	var resp struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse upstream response"})
		return
	}

	available := make([]gin.H, 0, len(resp.HeroEquipment))
	seen := map[string]struct{}{}
	for _, it := range resp.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
		available = append(available, gin.H{
			"name":      name,
			"level":     it.Level,
			"maxLevel":  it.MaxLevel,
			"available": true,
		})
	}

	unavailable := make([]gin.H, 0)
	for _, item := range h.catalog.Items {
		name := item.Name
		if _, ok := seen[name]; !ok {
			unavailable = append(unavailable, gin.H{
				"name":      name,
				"level":     0,
				"maxLevel":  0,
				"available": false,
			})
		}
	}

	sort.Slice(available, func(i, j int) bool { return available[i]["name"].(string) < available[j]["name"].(string) })
	sort.Slice(unavailable, func(i, j int) bool { return unavailable[i]["name"].(string) < unavailable[j]["name"].(string) })

	c.JSON(http.StatusOK, gin.H{
		"playerTag":   nTag,
		"available":   available,
		"unavailable": unavailable,
	})
}
