package config

import (
	"log"
	"os"
)

type Config struct {
	ServerAddr  string
	CocBaseURL  string
	CocAPIToken string
}

func Load() Config {
	cfg := Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		CocBaseURL:  getEnv("COC_API_BASE", "https://api.clashofclans.com/v1"),
		CocAPIToken: os.Getenv("COC_API_TOKEN"),
	}
	if cfg.CocAPIToken == "" {
		log.Println("warning: COC_API_TOKEN is not set; upstream calls will fail")
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
