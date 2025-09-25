package usecases

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"

	"github.com/dauletkhan/coc/internal/domain/ports"
)

type PlayerHeroEquipmentsUseCase struct {
	playerAPI ports.PlayerAPI
	catalog   ports.CatalogRepository
}

func NewPlayerHeroEquipmentsUseCase(playerAPI ports.PlayerAPI, catalog ports.CatalogRepository) *PlayerHeroEquipmentsUseCase {
	return &PlayerHeroEquipmentsUseCase{playerAPI: playerAPI, catalog: catalog}
}

type PlayerHeroEquipmentsResult struct {
	PlayerTag   string      `json:"playerTag"`
	Available   []Equipment `json:"available"`
	Unavailable []Equipment `json:"unavailable"`
}

type Equipment struct {
	Name      string `json:"name"`
	Level     int    `json:"level"`
	MaxLevel  int    `json:"maxLevel"`
	Available bool   `json:"available"`
}

func (uc *PlayerHeroEquipmentsUseCase) Execute(ctx context.Context, playerTag string) (PlayerHeroEquipmentsResult, int, error) {
	var out PlayerHeroEquipmentsResult
	// Validate path-safety for tag
	if _, err := url.PathUnescape(playerTag); err != nil {
		return out, 400, err
	}
	body, status, err := uc.playerAPI.GetPlayerRaw(ctx, playerTag)
	if err != nil || status >= 400 {
		return out, status, err
	}
	type equipment struct {
		Name     json.RawMessage `json:"name"`
		Level    int             `json:"level"`
		MaxLevel int             `json:"maxLevel"`
	}
	var resp struct {
		HeroEquipment []equipment `json:"heroEquipment"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return out, 500, err
	}
	available := make([]Equipment, 0, len(resp.HeroEquipment))
	seen := map[string]struct{}{}
	for _, it := range resp.HeroEquipment {
		name := extractName(it.Name)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
		available = append(available, Equipment{
			Name:      name,
			Level:     it.Level,
			MaxLevel:  it.MaxLevel,
			Available: true,
		})
	}
	unavailable := make([]Equipment, 0)
	// catalog-only names are considered unavailable
	for _, name := range uc.catalog.ListEquipmentNames() {
		if _, ok := seen[name]; !ok {
			unavailable = append(unavailable, Equipment{
				Name:      name,
				Level:     0,
				MaxLevel:  0,
				Available: false,
			})
		}
	}
	sort.Slice(available, func(i, j int) bool { return available[i].Name < available[j].Name })
	sort.Slice(unavailable, func(i, j int) bool { return unavailable[i].Name < unavailable[j].Name })
	out.PlayerTag = playerTag
	out.Available = available
	out.Unavailable = unavailable
	return out, 200, nil
}
