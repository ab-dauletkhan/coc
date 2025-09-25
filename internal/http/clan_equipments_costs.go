package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dauletkhan/coc/internal/catalog"
	"github.com/dauletkhan/coc/internal/coc"
)

type ClanEquipmentsCostsHandler struct {
	coc     *coc.Client
	catalog catalog.EquipmentCatalog
}

func NewClanEquipmentsCostsHandler(cocClient *coc.Client, cat catalog.EquipmentCatalog) *ClanEquipmentsCostsHandler {
	return &ClanEquipmentsCostsHandler{coc: cocClient, catalog: cat}
}

func (h *ClanEquipmentsCostsHandler) Register(r *gin.Engine) {
	r.GET("/v1/clans/:tag/hero-equipments/costs", h.getClanEquipmentsCosts)
}

type ore struct{ shiny, glowy, starry int }

type memberCost struct {
	Tag   string `json:"tag"`
	Name  string `json:"name"`
	Spent ore    `json:"spent"`
}

func (h *ClanEquipmentsCostsHandler) getClanEquipmentsCosts(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tag"})
		return
	}
	nTag := normalizePlayerTag(tag)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Fetch clan members
	b, status, err := h.coc.GetClanMembersRaw(ctx, nTag)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if status >= 400 {
		c.Data(status, "application/json", b)
		return
	}
	var members struct {
		Items []struct {
			Tag  string `json:"tag"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(b, &members); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse members"})
		return
	}

	// Concurrently fetch each player's costs
	workerLimit := 5
	sem := make(chan struct{}, workerLimit)
	wg := sync.WaitGroup{}
	results := make([]memberCost, len(members.Items))

	for i, m := range members.Items {
		i, m := i, m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			ctxp, cancelp := context.WithTimeout(ctx, 6*time.Second)
			defer cancelp()
			pb, pstatus, perr := h.coc.GetPlayerRaw(ctxp, normalizePlayerTag(m.Tag))
			spent := ore{}
			if perr == nil && pstatus < 400 {
				spent = h.computePlayerOre(pb)
			}
			results[i] = memberCost{Tag: m.Tag, Name: m.Name, Spent: spent}
		}()
	}
	wg.Wait()

	// Sort by shiny desc, then glowy desc, then starry desc
	sort.Slice(results, func(i, j int) bool {
		if results[i].Spent.shiny != results[j].Spent.shiny {
			return results[i].Spent.shiny > results[j].Spent.shiny
		}
		if results[i].Spent.glowy != results[j].Spent.glowy {
			return results[i].Spent.glowy > results[j].Spent.glowy
		}
		return results[i].Spent.starry > results[j].Spent.starry
	})

	// Totals
	tot := ore{}
	for _, r := range results {
		tot.shiny += r.Spent.shiny
		tot.glowy += r.Spent.glowy
		tot.starry += r.Spent.starry
	}

	// Build JSON
	resp := gin.H{
		"clanTag": nTag,
		"total":   gin.H{"shiny": tot.shiny, "glowy": tot.glowy, "starry": tot.starry},
		"members": func() []gin.H {
			out := make([]gin.H, 0, len(results))
			for _, r := range results {
				out = append(out, gin.H{
					"tag":   r.Tag,
					"name":  r.Name,
					"spent": gin.H{"shiny": r.Spent.shiny, "glowy": r.Spent.glowy, "starry": r.Spent.starry},
				})
			}
			return out
		}(),
	}
	c.JSON(http.StatusOK, resp)
}

// computePlayerOre parses a player payload and computes total ore spent using the catalog.
func (h *ClanEquipmentsCostsHandler) computePlayerOre(body []byte) ore {
	type equipment struct {
		Name  json.RawMessage `json:"name"`
		Level int             `json:"level"`
	}
	var p struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return ore{}
	}
	// Map catalog
	catByName := make(map[string]string, len(h.catalog.Items))
	for _, it := range h.catalog.Items {
		catByName[it.Name] = it.Rarity
	}
	shiny, glowy, starry := 0, 0, 0
	for _, it := range p.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		rarity, ok := catByName[name]
		if !ok {
			rarity = inferRarity(name)
		}
		lvl := it.Level
		if lvl < 0 {
			lvl = 0
		}
		var table []catalog.OreCost
		switch strings.ToUpper(rarity) {
		case "COMMON":
			table = h.catalog.CommonCostsPerLevel
		case "EPIC":
			table = h.catalog.EpicCostsPerLevel
		}
		if len(table) == 0 {
			continue
		}
		if lvl >= len(table) {
			lvl = len(table) - 1
		}
		for i := 0; i <= lvl; i++ {
			shiny += table[i].Shiny
			glowy += table[i].Glowy
			starry += table[i].Starry
		}
	}
	return ore{shiny: shiny, glowy: glowy, starry: starry}
}
