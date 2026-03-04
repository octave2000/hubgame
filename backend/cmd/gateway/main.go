package main

import (
	"log"
	"net/http"
	"time"

	"hubgame/backend/internal/api"
	"hubgame/backend/internal/platform"
)

func main() {
	cfg := platform.LoadConfig()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewGatewayServer(cfg.ControllerURL, cfg.DBEngineURL, cfg.InternalServiceToken, cfg.ControllerAdminToken, cfg.EnableDevAuth).Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("gateway listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
