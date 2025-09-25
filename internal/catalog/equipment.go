package catalog

import (
	"encoding/json"
	"errors"
	"os"
)

type EquipmentCatalog struct {
	Items               []Equipment `json:"items"`
	CommonCostsPerLevel []OreCost   `json:"commonCostsPerLevel"`
	EpicCostsPerLevel   []OreCost   `json:"epicCostsPerLevel"`
}

// Hero is an enum describing which hero an equipment belongs to.
type Hero string

const (
	HeroBarbarianKing Hero = "BARBARIAN_KING"
	HeroArcherQueen   Hero = "ARCHER_QUEEN"
	HeroGrandWarden   Hero = "GRAND_WARDEN"
	HeroRoyalChampion Hero = "ROYAL_CHAMPION"
	HeroMinionPrince  Hero = "MINION_PRINCE"
)

type Equipment struct {
	Name   string `json:"name"`
	Rarity string `json:"rarity"` // COMMON or EPIC
	Hero   Hero   `json:"hero"`   // e.g. BARBARIAN_KING, ARCHER_QUEEN, GRAND_WARDEN, ROYAL_CHAMPION, MINION_PRINCE
	ID     int    `json:"id"`     // sortable stable identifier
}

type OreCost struct {
	Shiny  int `json:"shiny"`
	Glowy  int `json:"glowy"`
	Starry int `json:"starry"`
}

func LoadEquipmentCatalog(path string) (EquipmentCatalog, error) {
	var cat EquipmentCatalog
	f, err := os.Open(path)
	if err != nil {
		return cat, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cat); err != nil {
		return cat, err
	}
	for i := range cat.Items {
		if cat.Items[i].Name == "" {
			return cat, errors.New("equipment catalog item missing name")
		}
	}
	return cat, nil
}
