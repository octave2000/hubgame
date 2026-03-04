package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"hubgame/backend/internal/database"
)

// SchemaController validates payload shapes for entities and events.
type SchemaController struct{}

func (SchemaController) Name() string { return "schema_controller" }

func (SchemaController) BeforeInsert(_ context.Context, e *database.Entity) error {
	return validateEntitySchema(e.Kind, e.Data)
}

func (SchemaController) BeforeUpdate(_ context.Context, _ *database.Entity, next *database.Entity) error {
	return validateEntitySchema(next.Kind, next.Data)
}

func (SchemaController) BeforeDelete(context.Context, *database.Entity) error {
	return nil
}

func (SchemaController) BeforeAppendEvent(_ context.Context, event *database.Event) error {
	return validateEventSchema(event.Type, event.Payload)
}

func validateEntitySchema(kind string, data json.RawMessage) error {
	requiredByKind := map[string][]string{
		"user":              {"username"},
		"room":              {"name"},
		"match":             {"mode", "status"},
		"game":              {"name", "studio", "category", "modes", "vibe", "cover"},
		"leaderboard_user":  {"user_id", "display_name"},
		"leaderboard_score": {"game_id", "user_id", "score", "hubcoins"},
		"tiktoe_match":      {"id", "mode", "board_size", "win_length", "board", "current"},
		"tiktoe_queue":      {"user_id", "display_name"},
		"tiktoe_chat":       {"id", "match_id", "user_id", "type"},
	}
	return validateRequiredFields("entity", kind, data, requiredByKind)
}

func validateEventSchema(eventType string, payload json.RawMessage) error {
	requiredByType := map[string][]string{
		"entity.inserted":             {"id"},
		"entity.updated":              {"id"},
		"entity.deleted":              {"id"},
		"match.create":                {"match_id", "mode"},
		"match.join":                  {"match_id", "user_id"},
		"move.place":                  {"match_id", "x", "y"},
		"chat.send":                   {"room_id", "message"},
		"reaction.send":               {"room_id", "emoji"},
		"leaderboard.score_submitted": {"game_id", "user_id", "score", "hubcoins"},
		"tiktoe.match_created":        {"id", "mode", "board_size", "win_length"},
		"tiktoe.match_found":          {"id", "player_x", "player_o"},
		"tiktoe.move":                 {"id", "board", "current"},
		"tiktoe.chat":                 {"id", "match_id", "user_id", "type"},
	}
	return validateRequiredFields("event", eventType, payload, requiredByType)
}

func validateRequiredFields(scope, key string, payload json.RawMessage, required map[string][]string) error {
	fields, ok := required[key]
	if !ok {
		return nil
	}
	if len(payload) == 0 {
		return fmt.Errorf("%s %q payload is empty", scope, key)
	}
	obj := map[string]any{}
	if err := json.Unmarshal(payload, &obj); err != nil {
		return errors.New("payload must be valid JSON object")
	}
	for _, field := range fields {
		if _, ok := obj[field]; !ok {
			return fmt.Errorf("%s %q is missing field %q", scope, key, field)
		}
	}
	return nil
}
