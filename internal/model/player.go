package model

// Player mirrors a subset of the Clash of Clans Player schema we need now.
type Player struct {
	HeroEquipment []PlayerItemLevel `json:"heroEquipment"`
}

type PlayerItemLevel struct {
	Name struct {
		LocalizedName string `json:"name"`
	} `json:"name"`
	Level    int    `json:"level"`
	MaxLevel int    `json:"maxLevel"`
	Village  string `json:"village"`
}
