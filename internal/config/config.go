package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	// Domains
	SupportedDomains []string

	// Database
	IsDBEnabled  bool
	DBDriver     string // "mysql", "mariadb", or "postgres"
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	QueryDomains string
	QueryUsers   string

	// Listening address
	ListenAddr string
}

// Load reads a .env file (if present) then overlays actual environment variables.
// It validates required fields and returns a populated Config or a fatal error.
func Load() (*Config, error) {
	// Load .env if present; ignore error if file is absent.
	_ = godotenv.Load()

	cfg := &Config{}

	// --- Domains ---
	raw := os.Getenv("SUPPORTED_DOMAINS")
	if raw != "" {
		for _, d := range strings.Split(raw, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				cfg.SupportedDomains = append(cfg.SupportedDomains, strings.ToLower(d))
			}
		}
	}

	// --- Database ---
	cfg.IsDBEnabled = strings.EqualFold(os.Getenv("ISDBENABLED"), "true")
	if cfg.IsDBEnabled {
		cfg.DBDriver = strings.ToLower(strings.TrimSpace(os.Getenv("DBDRIVER")))
		if cfg.DBDriver == "" {
			cfg.DBDriver = "mysql"
		}
		if cfg.DBDriver != "mysql" && cfg.DBDriver != "mariadb" && cfg.DBDriver != "postgres" {
			return nil, fmt.Errorf("unsupported DBDRIVER %q: must be mysql, mariadb, or postgres", cfg.DBDriver)
		}

		cfg.DBHost = envOrDefault("DBHOST", "localhost")
		cfg.DBName = os.Getenv("DBNAME")
		cfg.DBUser = os.Getenv("DBUSER")
		cfg.DBPassword = os.Getenv("DBPASSW")

		// Default port by driver
		switch cfg.DBDriver {
		case "postgres":
			cfg.DBPort = envOrDefault("DBPORT", "5432")
		default:
			cfg.DBPort = envOrDefault("DBPORT", "3306")
		}

		// Validate required DB fields
		if cfg.DBName == "" {
			return nil, fmt.Errorf("DBNAME is required when ISDBENABLED=true")
		}
		if cfg.DBUser == "" {
			return nil, fmt.Errorf("DBUSER is required when ISDBENABLED=true")
		}

		// Validate queries contain at least one placeholder
		cfg.QueryDomains = os.Getenv("QUERY_DOMAINS")
		cfg.QueryUsers = os.Getenv("QUERY_USERS")

		if cfg.QueryDomains == "" {
			return nil, fmt.Errorf("QUERY_DOMAINS is required when ISDBENABLED=true")
		}
		if cfg.QueryUsers == "" {
			return nil, fmt.Errorf("QUERY_USERS is required when ISDBENABLED=true")
		}
		if err := validatePlaceholders(cfg.DBDriver, "QUERY_DOMAINS", cfg.QueryDomains); err != nil {
			return nil, err
		}
		if err := validatePlaceholders(cfg.DBDriver, "QUERY_USERS", cfg.QueryUsers); err != nil {
			return nil, err
		}
	}

	// --- HTTP listener ---
	cfg.ListenAddr = envOrDefault("LISTEN_ADDR", ":8080")

	// Warn if neither domains nor DB are configured
	if !cfg.IsDBEnabled && len(cfg.SupportedDomains) == 0 {
		log.Println("WARNING: SUPPORTED_DOMAINS is empty and ISDBENABLED is false — all domain lookups will fail")
	}

	return cfg, nil
}

// validatePlaceholders ensures a user-supplied SQL query contains at least one
// positional placeholder, preventing raw string interpolation of user input.
func validatePlaceholders(driver, varName, query string) error {
	switch driver {
	case "postgres":
		if !strings.Contains(query, "$1") {
			return fmt.Errorf("%s must contain at least one positional placeholder ($1) for Postgres", varName)
		}
	default: // mysql / mariadb
		if !strings.Contains(query, "?") {
			return fmt.Errorf("%s must contain at least one positional placeholder (?) for MySQL/MariaDB", varName)
		}
	}
	return nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
