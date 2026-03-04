package seed

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"hubgame/backend/internal/database"
)

const CatalogSeedName = "catalog"
const CatalogSeedVersion = "2026-03-04-v1"

var CatalogTenantID = "hubgame-dev"

type CatalogGame struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Studio    string   `json:"studio"`
	Category  string   `json:"category"`
	Modes     []string `json:"modes"`
	Vibe      string   `json:"vibe"`
	Cover     string   `json:"cover"`
	Rating    float64  `json:"rating"`
	Players   string   `json:"players"`
	Installed bool     `json:"installed"`
	Tags      []string `json:"tags"`
}

func CatalogGames() []CatalogGame {
	return []CatalogGame{
		{
			ID:        "mod-grid",
			Name:      "Modifier Grid",
			Studio:    "Northline Labs",
			Category:  "Strategy",
			Modes:     []string{"Solo", "Online Multiplayer", "Bot Match"},
			Vibe:      "Sharp tactical rounds with playful modifiers.",
			Cover:     "https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.8,
			Players:   "42k online",
			Installed: true,
			Tags:      []string{"ranked", "modifiers", "pvp"},
		},
		{
			ID:        "rift-racers",
			Name:      "Rift Racers",
			Studio:    "Quiet Orbit",
			Category:  "Arcade",
			Modes:     []string{"Online Multiplayer", "Offline Multiplayer"},
			Vibe:      "Fast rounds, smooth drifts, bright rush moments.",
			Cover:     "https://images.unsplash.com/photo-1579373903781-fd5c0c30c4cd?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.6,
			Players:   "18k online",
			Installed: false,
			Tags:      []string{"racing", "party", "quick"},
		},
		{
			ID:        "tiny-kingdoms",
			Name:      "Tiny Kingdoms",
			Studio:    "Clay Beacon",
			Category:  "Strategy",
			Modes:     []string{"Solo", "Bot Match"},
			Vibe:      "Small maps with surprisingly deep decisions.",
			Cover:     "https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.7,
			Players:   "11k online",
			Installed: false,
			Tags:      []string{"simulation", "deep", "solo"},
		},
		{
			ID:        "echo-tiles",
			Name:      "Echo Tiles",
			Studio:    "Moss Arcade",
			Category:  "Puzzle",
			Modes:     []string{"Solo", "Online Multiplayer"},
			Vibe:      "Rhythm and puzzle flow for calm sessions.",
			Cover:     "https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.5,
			Players:   "9k online",
			Installed: true,
			Tags:      []string{"relaxing", "co-op", "focus"},
		},
		{
			ID:        "couch-club",
			Name:      "Couch Club",
			Studio:    "Gather House",
			Category:  "Party",
			Modes:     []string{"Offline Multiplayer", "Online Multiplayer"},
			Vibe:      "Light mini-games for local nights.",
			Cover:     "https://images.unsplash.com/photo-1511882150382-421056c89033?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.4,
			Players:   "24k online",
			Installed: false,
			Tags:      []string{"friends", "casual", "fun"},
		},
		{
			ID:        "canvas-quest",
			Name:      "Canvas Quest",
			Studio:    "Signal Pine",
			Category:  "Adventure",
			Modes:     []string{"Solo", "Online Multiplayer"},
			Vibe:      "Warm world exploration with cooperative moments.",
			Cover:     "https://images.unsplash.com/photo-1493711662062-fa541adb3fc8?auto=format&fit=crop&w=1800&q=80",
			Rating:    4.9,
			Players:   "31k online",
			Installed: false,
			Tags:      []string{"story", "co-op", "exploration"},
		},
	}
}

func ApplyCatalog(ctx context.Context, store *database.Store) error {
	for _, game := range CatalogGames() {
		data, err := json.Marshal(game)
		if err != nil {
			return err
		}
		next := &database.Entity{
			ID:       game.ID,
			TenantID: CatalogTenantID,
			Kind:     "game",
			Data:     data,
		}
		if err := store.InsertEntity(ctx, next); err != nil {
			if !isConflict(err) {
				return err
			}
			if err := store.UpdateEntity(ctx, &database.Entity{
				ID:       game.ID,
				TenantID: CatalogTenantID,
				Data:     data,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func isConflict(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, database.ErrVersionConflict) ||
		strings.Contains(err.Error(), "UNIQUE constraint failed")
}
