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

type Equipment struct {
	Name   string `json:"name"`
	Rarity string `json:"rarity"` // COMMON or EPIC
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
