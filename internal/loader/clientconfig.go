package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ClientConfig represents a vendor's autodiscovery metadata loaded from a
// JSON file in the clientConfigs/ directory.
type ClientConfig struct {
	// Vendor is a short identifier used to key TMPL_<VENDOR>_<KEY> env overrides.
	Vendor string `json:"vendor"`

	// TemplateFile is the path (relative to the service root) of the Go template
	// that will be rendered for requests matched by this config.
	TemplateFile string `json:"templateFile"`

	// SupportedEndpoints lists the URL paths this vendor config handles.
	SupportedEndpoints []string `json:"supportedEndpoints"`

	// Vars holds default template variables. Values may be overridden at runtime
	// via TMPL_<VENDOR>_<KEY> environment variables.
	Vars map[string]string `json:"vars"`
}

// LoadClientConfigs scans dir for all *.json files and returns the parsed configs.
// It returns an error if any file cannot be parsed or is missing required fields.
func LoadClientConfigs(dir string) ([]*ClientConfig, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("scanning clientConfigs dir: %w", err)
	}

	var configs []*ClientConfig
	for _, path := range entries {
		cfg, err := parseClientConfig(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func parseClientConfig(path string) (*ClientConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg ClientConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate required fields
	if cfg.Vendor == "" {
		return nil, fmt.Errorf("missing required field \"vendor\"")
	}
	if cfg.TemplateFile == "" {
		return nil, fmt.Errorf("missing required field \"templateFile\"")
	}
	if len(cfg.SupportedEndpoints) == 0 {
		return nil, fmt.Errorf("\"supportedEndpoints\" must contain at least one entry")
	}

	// Normalise vendor to uppercase for consistent env-var key lookups
	cfg.Vendor = strings.ToUpper(cfg.Vendor)

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]string)
	}

	// Apply TMPL_<VENDOR>_<KEY> environment variable overrides
	prefix := "TMPL_" + cfg.Vendor + "_"
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		kv := strings.SplitN(env, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimPrefix(kv[0], prefix)
		cfg.Vars[key] = kv[1]
	}

	return &cfg, nil
}
