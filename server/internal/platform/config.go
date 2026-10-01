package platform

import (
	"os"
)

type Config struct {
	Port         string
	DBPath       string
	GatewayID    string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "messaging.db"
	}

	gatewayID := os.Getenv("GATEWAY_ID")
	if gatewayID == "" {
		gatewayID = "gw-local-1"
	}

	return Config{
		Port:         port,
		DBPath:       dbPath,
		GatewayID:    gatewayID,
	}
}
