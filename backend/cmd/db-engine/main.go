package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"hubgame/backend/internal/api"
	"hubgame/backend/internal/controller"
	"hubgame/backend/internal/database"
	"hubgame/backend/internal/platform"
)

func main() {
	cfg := platform.LoadConfig()
	ctx := context.Background()

	store, err := database.OpenSQLite(ctx, cfg.SQLiteDSN)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	store.RegisterController(controller.SchemaController{})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewDBEngineServer(store, cfg.InternalServiceToken).Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("db-engine listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
