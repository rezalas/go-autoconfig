package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"go-autoconfig/internal/config"
	"go-autoconfig/internal/db"
	"go-autoconfig/internal/handler"
	"go-autoconfig/internal/loader"
	"go-autoconfig/internal/validate"
)

func main() {
	// --- Configuration ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// --- Resolve paths relative to executable ---
	baseDir := execDir()
	clientConfigsDir := filepath.Join(baseDir, "clientConfigs")
	templatesDir := filepath.Join(baseDir, "templates")

	// --- Load client configs ---
	clientConfigs, err := loader.LoadClientConfigs(clientConfigsDir)
	if err != nil {
		log.Fatalf("loading clientConfigs: %v", err)
	}
	if len(clientConfigs) == 0 {
		log.Fatal("no clientConfigs found — add at least one JSON file to clientConfigs/")
	}
	log.Printf("loaded %d client config(s)", len(clientConfigs))

	// --- Load templates ---
	registry, err := loader.LoadTemplates(templatesDir)
	if err != nil {
		log.Fatalf("loading templates: %v", err)
	}
	log.Printf("loaded %d template(s)", len(registry))

	// --- Domain checker ---
	var checker validate.DomainChecker
	if cfg.IsDBEnabled {
		database, err := db.Open(cfg)
		if err != nil {
			log.Fatalf("connecting to database: %v", err)
		}
		defer database.Close()
		checker = database
		log.Printf("database validation enabled (%s)", cfg.DBDriver)
	} else {
		checker = validate.NewEnvChecker(cfg.SupportedDomains)
		log.Printf("env-var validation enabled (%d domain(s))", len(cfg.SupportedDomains))
	}

	// --- HTTP server ---
	r := gin.Default()
	handler.RegisterRoutes(r, clientConfigs, registry, checker)

	log.Printf("listening on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// execDir returns the directory containing the running executable, falling
// back to the current directory if it cannot be determined.
func execDir() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}
