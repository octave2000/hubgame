package main

import (
	"log"
	"net/http"
	"time"

	"hubgame/backend/internal/api"
	"hubgame/backend/internal/controller"
	"hubgame/backend/internal/platform"
)

func main() {
	cfg := platform.LoadConfig()
	auth := controller.NewAuthController(cfg.JWTSecret, cfg.Issuer)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewControllerService(auth, cfg.ControllerAdminToken).Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("controller listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
