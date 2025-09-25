package ports

import "context"

// PlayerAPI defines secondary port for fetching player data from an external service.
type PlayerAPI interface {
	GetPlayerRaw(ctx context.Context, tag string) ([]byte, int, error)
}

// ClanAPI defines secondary port for fetching clan data from an external service.
type ClanAPI interface {
	GetClanMembersRaw(ctx context.Context, tag string) ([]byte, int, error)
}
