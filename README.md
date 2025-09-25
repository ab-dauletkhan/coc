# Custom Clash of Clans API (Go)

A small Go service that wraps the official Clash of Clans API and adds
player hero equipment utilities (availability and cumulative ore costs).

## Prerequisites
- Go 1.22+
- Clash of Clans API token

## Quick start
1. Create a `.env` file:
```
COC_API_TOKEN=your_token_here
# optional overrides
# SERVER_ADDR=:8080
# EQUIPMENT_CATALOG_PATH=data/hero_equipment.json
```
2. Install deps and run:
```
go mod tidy
go run ./cmd/api
```
3. Open Swagger UI:
- http://localhost:8080/docs

## Endpoints
- GET `/v1/players/{tag}/hero-equipments`
  - Lists player's hero equipments and marks missing ones as unavailable.
  - Tag can be `#ABC123`, `%23ABC123`, or `ABC123`.

- GET `/v1/players/{tag}/hero-equipments/costs`
  - Computes cumulative ore spent per equipment and totals.
  - Uses per-rarity per-level costs from `data/hero_equipment.json`.

## Catalog data
The service reads equipment names/rarities and ore cost tables from:
- `data/hero_equipment.json`

This file is not sourced from the official API and should be maintained
manually. Update names and rarities as needed and restart the service.

## Notes
- The service uses the official API only to fetch the player payload.
- Rate limits and error codes from the upstream API are proxied.
- Production hardening: add Redis caching, retries/backoff, and auth.
