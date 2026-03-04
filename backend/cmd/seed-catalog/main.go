package main

import (
	"context"
	"flag"
	"log"

	"hubgame/backend/internal/controller"
	"hubgame/backend/internal/database"
	"hubgame/backend/internal/platform"
	"hubgame/backend/internal/seed"
)

func main() {
	force := flag.Bool("force", false, "apply seed even if version already marked")
	flag.Parse()

	cfg := platform.LoadConfig()
	ctx := context.Background()

	store, err := database.OpenSQLite(ctx, cfg.SQLiteDSN)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()
	store.RegisterController(controller.SchemaController{})

	applied, err := store.IsSeedApplied(ctx, seed.CatalogSeedName, seed.CatalogSeedVersion)
	if err != nil {
		log.Fatalf("check seed history: %v", err)
	}
	if applied && !*force {
		log.Printf("seed %s@%s already applied; use -force to reapply", seed.CatalogSeedName, seed.CatalogSeedVersion)
		return
	}

	if err := seed.ApplyCatalog(ctx, store); err != nil {
		log.Fatalf("apply catalog seed: %v", err)
	}
	if err := store.MarkSeedApplied(ctx, seed.CatalogSeedName, seed.CatalogSeedVersion); err != nil {
		log.Fatalf("mark seed applied: %v", err)
	}

	log.Printf("seed applied successfully: %s@%s", seed.CatalogSeedName, seed.CatalogSeedVersion)
}
