package platform

import (
	"os"
	"strings"
)

type Config struct {
	Addr                 string
	SQLiteDSN            string
	JWTSecret            string
	Issuer               string
	ControllerURL        string
	DBEngineURL          string
	InternalServiceToken string
	ControllerAdminToken string
	EnableDevAuth        bool
}

func LoadConfig() Config {
	cfg := Config{
		Addr:                 envOr("HUBGAME_ADDR", ":8080"),
		SQLiteDSN:            envOr("HUBGAME_SQLITE_DSN", "file:hubgame.db?_pragma=busy_timeout(5000)"),
		JWTSecret:            envOr("HUBGAME_JWT_SECRET", "dev-secret-change-me"),
		Issuer:               envOr("HUBGAME_JWT_ISSUER", "hubgame-controller"),
		ControllerURL:        envOr("HUBGAME_CONTROLLER_URL", "http://controller:8082"),
		DBEngineURL:          envOr("HUBGAME_DB_ENGINE_URL", "http://db-engine:8081"),
		InternalServiceToken: envOr("HUBGAME_INTERNAL_TOKEN", "dev-internal-token"),
		ControllerAdminToken: envOr("HUBGAME_CONTROLLER_ADMIN_TOKEN", "dev-controller-admin"),
		EnableDevAuth:        envBool("HUBGAME_ENABLE_DEV_AUTH", true),
	}
	return cfg
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
