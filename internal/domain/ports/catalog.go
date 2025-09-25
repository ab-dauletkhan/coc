package ports

// CatalogRepository is a secondary port for accessing the equipment catalog.
type CatalogRepository interface {
	// GetRarity returns the rarity for the given equipment name or empty when unknown.
	GetRarity(name string) string
	// CostsCommon returns the per-level common costs.
	CostsCommon() []OreCost
	// CostsEpic returns the per-level epic costs.
	CostsEpic() []OreCost
	// ListEquipmentNames returns known equipment names in the catalog.
	ListEquipmentNames() []string
}

// OreCost is a small value object used by the catalog port.
type OreCost struct {
	Shiny  int
	Glowy  int
	Starry int
}
