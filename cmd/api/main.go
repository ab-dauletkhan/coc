package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ab-dauletkhan/coc/internal/catalog"
	"github.com/ab-dauletkhan/coc/internal/coc"
	"github.com/ab-dauletkhan/coc/internal/config"

	primaryhttp "github.com/ab-dauletkhan/coc/internal/adapters/primary/http"
	secondary "github.com/ab-dauletkhan/coc/internal/adapters/secondary"
	"github.com/ab-dauletkhan/coc/internal/application/usecases"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// Try to detect public outbound IP for allowlisting
	if ip := fetchPublicIP(); ip != "" {
		log.Printf("public outbound IP (for allowlist): %s", ip)
	} else {
		log.Println("public outbound IP not detected (network may block metadata services)")
	}

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
	// Hexagonal handlers
	catalogAdapter := secondary.NewCatalogAdapter(cat)
	cocAdapter := secondary.NewCocAPIAdapter(cocClient)
	playerCostsUC := usecases.NewPlayerEquipmentCostsUseCase(cocAdapter, catalogAdapter)
	playerCostsHandler := primaryhttp.NewPlayerEquipmentCostsHandler(playerCostsUC)
	playerCostsHandler.Register(r)

	playerEquipUC := usecases.NewPlayerHeroEquipmentsUseCase(cocAdapter, catalogAdapter)
	playerEquipHandler := primaryhttp.NewPlayerHeroEquipmentsHandler(playerEquipUC)
	playerEquipHandler.Register(r)

	clanCostsUC := usecases.NewClanEquipmentCostsUseCase(cocAdapter, cocAdapter, catalogAdapter)
	clanCostsHandler := primaryhttp.NewClanEquipmentCostsHandler(clanCostsUC)
	clanCostsHandler.Register(r)

	// Swagger UI & spec
	primaryhttp.RegisterSwagger(r)

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

func fetchPublicIP() string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://ifconfig.me",
		"https://icanhazip.com",
	}
	client := &http.Client{Timeout: 2 * time.Second}
	for _, url := range endpoints {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode >= 400 {
			continue
		}
		ip := strings.TrimSpace(string(b))
		if ip != "" {
			return ip
		}
	}
	return ""
}
