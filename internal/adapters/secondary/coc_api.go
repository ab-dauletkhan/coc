package secondary

import (
	"context"

	"github.com/dauletkhan/coc/internal/coc"
)

// CocAPIAdapter adapts internal/coc.Client to domain ports.
type CocAPIAdapter struct {
	client *coc.Client
}

func NewCocAPIAdapter(client *coc.Client) *CocAPIAdapter { return &CocAPIAdapter{client: client} }

func (a *CocAPIAdapter) GetPlayerRaw(ctx context.Context, tag string) ([]byte, int, error) {
	return a.client.GetPlayerRaw(ctx, tag)
}

func (a *CocAPIAdapter) GetClanMembersRaw(ctx context.Context, tag string) ([]byte, int, error) {
	return a.client.GetClanMembersRaw(ctx, tag)
}
