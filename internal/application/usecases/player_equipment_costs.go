package usecases

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/dauletkhan/coc/internal/domain/models"
	"github.com/dauletkhan/coc/internal/domain/ports"
)

// PlayerEquipmentCostsUseCase computes per-equipment and total ore spent for a player.
type PlayerEquipmentCostsUseCase struct {
	playerAPI ports.PlayerAPI
	catalog   ports.CatalogRepository
}

func NewPlayerEquipmentCostsUseCase(playerAPI ports.PlayerAPI, catalog ports.CatalogRepository) *PlayerEquipmentCostsUseCase {
	return &PlayerEquipmentCostsUseCase{playerAPI: playerAPI, catalog: catalog}
}

type EquipmentSpend struct {
	Name   string           `json:"name"`
	Rarity string           `json:"rarity"`
	Level  int              `json:"level"`
	Spent  models.OreTotals `json:"spent"`
}

type PlayerEquipmentCostsResult struct {
	PlayerTag  string           `json:"playerTag"`
	Total      models.OreTotals `json:"total"`
	Equipments []EquipmentSpend `json:"equipments"`
}

func (uc *PlayerEquipmentCostsUseCase) Execute(ctx context.Context, playerTag string) (PlayerEquipmentCostsResult, int, error) {
	var out PlayerEquipmentCostsResult

	body, status, err := uc.playerAPI.GetPlayerRaw(ctx, playerTag)
	if err != nil || status >= 400 {
		return out, status, err
	}
	type equipment struct {
		Name  json.RawMessage `json:"name"`
		Level int             `json:"level"`
	}
	var resp struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return out, 500, err
	}

	catCommon := uc.catalog.CostsCommon()
	catEpic := uc.catalog.CostsEpic()

	var total models.OreTotals
	results := make([]EquipmentSpend, 0, len(resp.HeroEquipment))
	for _, it := range resp.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		rarity := uc.catalog.GetRarity(name)
		if rarity == "" {
			// Unknown in catalog, skip from cost computation as we cannot determine table
			continue
		}
		lvl := it.Level
		if lvl < 0 {
			lvl = 0
		}
		var table []ports.OreCost
		switch strings.ToUpper(rarity) {
		case "COMMON":
			table = catCommon
		case "EPIC":
			table = catEpic
		default:
			table = nil
		}
		spent := models.OreTotals{}
		if len(table) > 0 {
			if lvl >= len(table) {
				lvl = len(table) - 1
			}
			for i := 0; i <= lvl; i++ {
				spent.Shiny += table[i].Shiny
				spent.Glowy += table[i].Glowy
				spent.Starry += table[i].Starry
			}
		}
		total.Shiny += spent.Shiny
		total.Glowy += spent.Glowy
		total.Starry += spent.Starry
		results = append(results, EquipmentSpend{
			Name:   name,
			Rarity: strings.ToUpper(rarity),
			Level:  it.Level,
			Spent:  spent,
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	out.PlayerTag = playerTag
	out.Total = total
	out.Equipments = results
	return out, 200, nil
}

// helpers are provided by helpers.go in this package
