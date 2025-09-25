package usecases

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dauletkhan/coc/internal/domain/models"
	"github.com/dauletkhan/coc/internal/domain/ports"
)

type ClanEquipmentCostsUseCase struct {
	clanAPI   ports.ClanAPI
	playerAPI ports.PlayerAPI
	catalog   ports.CatalogRepository
}

func NewClanEquipmentCostsUseCase(clanAPI ports.ClanAPI, playerAPI ports.PlayerAPI, catalog ports.CatalogRepository) *ClanEquipmentCostsUseCase {
	return &ClanEquipmentCostsUseCase{clanAPI: clanAPI, playerAPI: playerAPI, catalog: catalog}
}

type ClanMemberSpend struct {
	Tag   string           `json:"tag"`
	Name  string           `json:"name"`
	Spent models.OreTotals `json:"spent"`
}

type ClanEquipmentCostsResult struct {
	ClanTag string            `json:"clanTag"`
	Total   models.OreTotals  `json:"total"`
	Members []ClanMemberSpend `json:"members"`
}

func (uc *ClanEquipmentCostsUseCase) Execute(ctx context.Context, clanTag string) (ClanEquipmentCostsResult, int, error) {
	var out ClanEquipmentCostsResult

	b, status, err := uc.clanAPI.GetClanMembersRaw(ctx, clanTag)
	if err != nil || status >= 400 {
		return out, status, err
	}
	var members struct {
		Items []struct {
			Tag  string `json:"tag"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(b, &members); err != nil {
		return out, 500, err
	}

	workerLimit := 5
	sem := make(chan struct{}, workerLimit)
	wg := sync.WaitGroup{}
	results := make([]ClanMemberSpend, len(members.Items))

	for i, m := range members.Items {
		i, m := i, m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			ctxp, cancelp := context.WithTimeout(ctx, 6*time.Second)
			defer cancelp()
			pb, pstatus, perr := uc.playerAPI.GetPlayerRaw(ctxp, normalizePlayerTag(m.Tag))
			spent := models.OreTotals{}
			if perr == nil && pstatus < 400 {
				spent = uc.computePlayerOre(pb)
			}
			results[i] = ClanMemberSpend{Tag: m.Tag, Name: m.Name, Spent: spent}
		}()
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Spent.Shiny != results[j].Spent.Shiny {
			return results[i].Spent.Shiny > results[j].Spent.Shiny
		}
		if results[i].Spent.Glowy != results[j].Spent.Glowy {
			return results[i].Spent.Glowy > results[j].Spent.Glowy
		}
		return results[i].Spent.Starry > results[j].Spent.Starry
	})

	tot := models.OreTotals{}
	for _, r := range results {
		tot.Shiny += r.Spent.Shiny
		tot.Glowy += r.Spent.Glowy
		tot.Starry += r.Spent.Starry
	}

	out.ClanTag = clanTag
	out.Total = tot
	out.Members = results
	return out, 200, nil
}

func (uc *ClanEquipmentCostsUseCase) computePlayerOre(body []byte) models.OreTotals {
	type equipment struct {
		Name  json.RawMessage `json:"name"`
		Level int             `json:"level"`
	}
	var p struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return models.OreTotals{}
	}
	shiny, glowy, starry := 0, 0, 0
	common := uc.catalog.CostsCommon()
	epic := uc.catalog.CostsEpic()
	for _, it := range p.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		rarity := uc.catalog.GetRarity(name)
		if rarity == "" {
			// Unknown in catalog
			continue
		}
		lvl := it.Level
		if lvl < 0 {
			lvl = 0
		}
		var table []ports.OreCost
		switch strings.ToUpper(rarity) {
		case "COMMON":
			table = common
		case "EPIC":
			table = epic
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
	return models.OreTotals{Shiny: shiny, Glowy: glowy, Starry: starry}
}

// inferRarity is provided by helpers.go in this package
