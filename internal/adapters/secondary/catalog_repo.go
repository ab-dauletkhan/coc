package secondary

import (
	"strings"

	"github.com/dauletkhan/coc/internal/catalog"
	"github.com/dauletkhan/coc/internal/domain/ports"
)

// CatalogAdapter adapts internal/catalog to the domain CatalogRepository port.
type CatalogAdapter struct {
	cat catalog.EquipmentCatalog
}

func NewCatalogAdapter(cat catalog.EquipmentCatalog) *CatalogAdapter {
	return &CatalogAdapter{cat: cat}
}

func (a *CatalogAdapter) GetRarity(name string) string {
	upper := strings.ToUpper(strings.TrimSpace(name))
	for _, it := range a.cat.Items {
		if strings.ToUpper(it.Name) == upper {
			return it.Rarity
		}
	}
	return ""
}

func (a *CatalogAdapter) CostsCommon() []ports.OreCost {
	out := make([]ports.OreCost, len(a.cat.CommonCostsPerLevel))
	for i, c := range a.cat.CommonCostsPerLevel {
		out[i] = ports.OreCost{Shiny: c.Shiny, Glowy: c.Glowy, Starry: c.Starry}
	}
	return out
}

func (a *CatalogAdapter) CostsEpic() []ports.OreCost {
	out := make([]ports.OreCost, len(a.cat.EpicCostsPerLevel))
	for i, c := range a.cat.EpicCostsPerLevel {
		out[i] = ports.OreCost{Shiny: c.Shiny, Glowy: c.Glowy, Starry: c.Starry}
	}
	return out
}

func (a *CatalogAdapter) ListEquipmentNames() []string {
	out := make([]string, 0, len(a.cat.Items))
	for _, it := range a.cat.Items {
		out = append(out, it.Name)
	}
	return out
}
