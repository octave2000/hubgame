package controller

import (
	"encoding/json"
	"testing"

	"hubgame/backend/internal/database"
)

func TestSchemaControllerEntityValidation(t *testing.T) {
	c := SchemaController{}

	err := c.BeforeInsert(nil, &database.Entity{
		Kind: "match",
		Data: json.RawMessage(`{"mode":"pvp"}`),
	})
	if err == nil {
		t.Fatalf("expected missing required field error")
	}

	err = c.BeforeInsert(nil, &database.Entity{
		Kind: "match",
		Data: json.RawMessage(`{"mode":"pvp","status":"created"}`),
	})
	if err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
}
