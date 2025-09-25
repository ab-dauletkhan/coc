package models

// Equipment represents a hero equipment known by the domain.
// This type is independent from any external API representation.
type Equipment struct {
	Name   string
	Rarity string // COMMON, EPIC, UNKNOWN
}

// EquipmentLevel captures a player's equipment level details.
type EquipmentLevel struct {
	Name     string
	Level    int
	MaxLevel int
}
