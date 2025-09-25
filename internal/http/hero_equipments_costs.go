package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dauletkhan/coc/internal/catalog"
	"github.com/dauletkhan/coc/internal/coc"
)

type HeroEquipmentsCostsHandler struct {
	coc     *coc.Client
	catalog catalog.EquipmentCatalog
}

func NewHeroEquipmentsCostsHandler(cocClient *coc.Client, cat catalog.EquipmentCatalog) *HeroEquipmentsCostsHandler {
	return &HeroEquipmentsCostsHandler{coc: cocClient, catalog: cat}
}

func (h *HeroEquipmentsCostsHandler) Register(r *gin.Engine) {
	r.GET("/v1/players/:tag/hero-equipments/costs", h.getHeroEquipmentsCosts)
}

func inferRarity(name string) string {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if strings.Contains(upper, "PUPPET") || upper == "GIANT ARROW" || upper == "HEALING TOME" || upper == "INVISIBILITY VIAL" || upper == "MAGIC MIRROR" {
		return "EPIC"
	}
	return "COMMON"
}

func (h *HeroEquipmentsCostsHandler) getHeroEquipmentsCosts(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tag"})
		return
	}
	nTag := normalizePlayerTag(tag)

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

	type equipment struct {
		Name  json.RawMessage `json:"name"`
		Level int             `json:"level"`
	}
	var resp struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse upstream response"})
		return
	}

	// Build a map of catalog by name
	catByName := make(map[string]catalog.Equipment, len(h.catalog.Items))
	for _, it := range h.catalog.Items {
		catByName[it.Name] = it
	}

	results := make([]gin.H, 0, len(resp.HeroEquipment))
	totalShiny, totalGlowy, totalStarry := 0, 0, 0
	for _, it := range resp.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		catItem, ok := catByName[name]
		rarity := "UNKNOWN"
		if ok {
			rarity = strings.ToUpper(catItem.Rarity)
		} else {
			rarity = inferRarity(name)
		}

		lvl := it.Level
		if lvl < 0 {
			lvl = 0
		}
		// choose cost table
		var table []catalog.OreCost
		switch rarity {
		case "COMMON":
			table = h.catalog.CommonCostsPerLevel
		case "EPIC":
			table = h.catalog.EpicCostsPerLevel
		default:
			table = nil
		}
		shiny, glowy, starry := 0, 0, 0
		if len(table) > 0 {
			if lvl >= len(table) {
				lvl = len(table) - 1
			}
			for i := 0; i <= lvl; i++ {
				shiny += table[i].Shiny
				glowy += table[i].Glowy
				starry += table[i].Starry
			}
		}
		totalShiny += shiny
		totalGlowy += glowy
		totalStarry += starry
		results = append(results, gin.H{
			"name":   name,
			"rarity": rarity,
			"level":  it.Level,
			"spent":  gin.H{"shiny": shiny, "glowy": glowy, "starry": starry},
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i]["name"].(string) < results[j]["name"].(string) })

	c.JSON(http.StatusOK, gin.H{
		"playerTag": nTag,
		"total": gin.H{
			"shiny":  totalShiny,
			"glowy":  totalGlowy,
			"starry": totalStarry,
		},
		"equipments": results,
	})
}
