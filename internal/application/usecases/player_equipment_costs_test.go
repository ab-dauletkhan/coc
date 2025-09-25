package usecases

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dauletkhan/coc/internal/domain/ports"
)

type stubPlayerAPI struct {
	body   []byte
	status int
	err    error
}

func (s stubPlayerAPI) GetPlayerRaw(ctx context.Context, tag string) ([]byte, int, error) {
	return s.body, s.status, s.err
}

type stubCatalog struct {
	names  []string
	rarity map[string]string
	common []struct{ Shiny, Glowy, Starry int }
	epic   []struct{ Shiny, Glowy, Starry int }
}

func (s stubCatalog) GetRarity(name string) string { return s.rarity[name] }
func (s stubCatalog) CostsCommon() []ports.OreCost {
	out := make([]ports.OreCost, len(s.common))
	for i, c := range s.common {
		out[i] = ports.OreCost{Shiny: c.Shiny, Glowy: c.Glowy, Starry: c.Starry}
	}
	return out
}
func (s stubCatalog) CostsEpic() []ports.OreCost {
	out := make([]ports.OreCost, len(s.epic))
	for i, c := range s.epic {
		out[i] = ports.OreCost{Shiny: c.Shiny, Glowy: c.Glowy, Starry: c.Starry}
	}
	return out
}
func (s stubCatalog) ListEquipmentNames() []string { return s.names }

func TestPlayerEquipmentCosts_Simple(t *testing.T) {
	payload := map[string]any{
		"heroEquipment": []map[string]any{
			{"name": "Rage Vial", "level": 2},
			{"name": "Giant Arrow", "level": 1},
		},
	}
	b, _ := json.Marshal(payload)
	uc := NewPlayerEquipmentCostsUseCase(
		stubPlayerAPI{body: b, status: 200},
		stubCatalog{
			names:  []string{"Rage Vial", "Giant Arrow"},
			rarity: map[string]string{"Rage Vial": "COMMON", "Giant Arrow": "EPIC"},
			common: []struct{ Shiny, Glowy, Starry int }{{10, 1, 0}, {20, 2, 0}, {30, 3, 0}},
			epic:   []struct{ Shiny, Glowy, Starry int }{{100, 10, 1}, {200, 20, 2}},
		},
	)
	res, status, err := uc.Execute(context.Background(), "%23TAG")
	if err != nil || status != 200 {
		t.Fatalf("unexpected err/status: %v %d", err, status)
	}
	if res.Total.Shiny != (10 + 20 + 100) {
		t.Fatalf("unexpected shiny total: %d", res.Total.Shiny)
	}
	if len(res.Equipments) != 2 {
		t.Fatalf("unexpected equipments len: %d", len(res.Equipments))
	}
}
