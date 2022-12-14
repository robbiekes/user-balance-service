package main

import (
	log "github.com/sirupsen/logrus"
	"user-balance-service/config"
	"user-balance-service/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.SetLogrus(cfg.Log.Level)

	// Run
	app.Run(cfg)
}
