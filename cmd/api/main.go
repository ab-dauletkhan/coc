package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/dauletkhan/coc/internal/catalog"
	"github.com/dauletkhan/coc/internal/coc"
	"github.com/dauletkhan/coc/internal/config"
	apphttp "github.com/dauletkhan/coc/internal/http"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	r := gin.Default()
	_ = r.SetTrustedProxies(nil)

	// Health check
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	cocClient := coc.NewClient(cfg.CocBaseURL, cfg.CocAPIToken)

	catalogPath := getenv("EQUIPMENT_CATALOG_PATH", "data/hero_equipment.json")
	cat, err := catalog.LoadEquipmentCatalog(catalogPath)
	if err != nil {
		log.Printf("warning: failed to load catalog at %s: %v", catalogPath, err)
	}

	// Handlers
	equipHandler := apphttp.NewHeroEquipmentsHandler(cocClient, cat)
	equipHandler.Register(r)
	costsHandler := apphttp.NewHeroEquipmentsCostsHandler(cocClient, cat)
	costsHandler.Register(r)

	// Swagger UI & spec
	apphttp.RegisterSwagger(r)

	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
