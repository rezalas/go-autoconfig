package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTemplates_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	registry, err := LoadTemplates(tmpDir)
	if err != nil {
		t.Fatalf("LoadTemplates() failed: %v", err)
	}

	if len(registry) != 0 {
		t.Errorf("expected empty registry for empty directory, got %d templates", len(registry))
	}
}

func TestLoadTemplates_SingleTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test template file
	templatePath := filepath.Join(tmpDir, "test.tmpl")
	err := os.WriteFile(templatePath, []byte("Hello {{.Name}}"), 0644)
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	registry, err := LoadTemplates(tmpDir)
	if err != nil {
		t.Fatalf("LoadTemplates() failed: %v", err)
	}

	if len(registry) != 1 {
		t.Errorf("expected 1 template, got %d", len(registry))
	}

	// Just verify a template exists, the exact key depends on the path
	for key, tmpl := range registry {
		if tmpl == nil {
			t.Errorf("template at key '%s' is nil", key)
		}
		if !strings.HasSuffix(key, "/test.tmpl") {
			t.Errorf("expected key to end with '/test.tmpl', got '%s'", key)
		}
	}
}

func TestLoadTemplates_InvalidTemplateFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid template file (with syntax error)
	templatePath := filepath.Join(tmpDir, "bad.tmpl")
	err := os.WriteFile(templatePath, []byte("{{.Unclosed"), 0644)
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	_, err = LoadTemplates(tmpDir)
	if err == nil {
		t.Errorf("expected error for invalid template, got nil")
	}
}

func TestTemplateRegistry_Get_Found(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := filepath.Join(tmpDir, "test.tmpl")
	err := os.WriteFile(templatePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	registry, err := LoadTemplates(tmpDir)
	if err != nil {
		t.Fatalf("LoadTemplates() failed: %v", err)
	}

	// Find the actual key in the registry
	var tmpl interface{}
	var actualKey string
	for k, v := range registry {
		actualKey = k
		tmpl = v
		break
	}

	if actualKey == "" {
		t.Fatalf("registry is empty, expected to find a template")
	}

	retrievedTmpl, err := registry.Get(actualKey)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrievedTmpl == nil {
		t.Errorf("expected template, got nil")
	}

	if retrievedTmpl != tmpl {
		t.Errorf("Get() returned different template than stored")
	}
}

func TestTemplateRegistry_Get_NotFound(t *testing.T) {
	registry := make(TemplateRegistry)

	_, err := registry.Get("/templates/nonexistent.tmpl")
	if err == nil {
		t.Errorf("expected error for missing template, got nil")
	}
}

func getRegistryKeys(registry TemplateRegistry) []string {
	var keys []string
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}

func TestParseClientConfig_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.json")
	configContent := `{
		"vendor": "Mozilla",
		"templateFile": "/templates/autoconfig.xml.tmpl",
		"supportedEndpoints": ["/mail/config-v1.1.xml"],
		"vars": {
			"HOSTNAME": "mail.example.com"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := parseClientConfig(configPath)
	if err != nil {
		t.Fatalf("parseClientConfig() failed: %v", err)
	}

	if cfg.Vendor != "MOZILLA" {
		t.Errorf("expected vendor 'MOZILLA', got '%s'", cfg.Vendor)
	}

	if cfg.TemplateFile != "/templates/autoconfig.xml.tmpl" {
		t.Errorf("expected templateFile '/templates/autoconfig.xml.tmpl', got '%s'", cfg.TemplateFile)
	}

	if len(cfg.SupportedEndpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(cfg.SupportedEndpoints))
	}

	if cfg.Vars["HOSTNAME"] != "mail.example.com" {
		t.Errorf("expected HOSTNAME='mail.example.com', got '%s'", cfg.Vars["HOSTNAME"])
	}
}

func TestParseClientConfig_MissingVendor(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.json")
	configContent := `{
		"templateFile": "/templates/autoconfig.xml.tmpl",
		"supportedEndpoints": ["/mail/config-v1.1.xml"]
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err = parseClientConfig(configPath)
	if err == nil {
		t.Errorf("expected error for missing vendor, got nil")
	}
}

func TestParseClientConfig_MissingTemplateFile(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.json")
	configContent := `{
		"vendor": "Mozilla",
		"supportedEndpoints": ["/mail/config-v1.1.xml"]
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err = parseClientConfig(configPath)
	if err == nil {
		t.Errorf("expected error for missing templateFile, got nil")
	}
}

func TestParseClientConfig_EmptyEndpoints(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.json")
	configContent := `{
		"vendor": "Mozilla",
		"templateFile": "/templates/autoconfig.xml.tmpl",
		"supportedEndpoints": []
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err = parseClientConfig(configPath)
	if err == nil {
		t.Errorf("expected error for empty endpoints, got nil")
	}
}

func TestParseClientConfig_EnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.json")
	configContent := `{
		"vendor": "Mozilla",
		"templateFile": "/templates/autoconfig.xml.tmpl",
		"supportedEndpoints": ["/mail/config-v1.1.xml"],
		"vars": {
			"HOSTNAME": "default.example.com"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set environment override
	os.Setenv("TMPL_MOZILLA_HOSTNAME", "override.example.com")
	defer os.Unsetenv("TMPL_MOZILLA_HOSTNAME")

	cfg, err := parseClientConfig(configPath)
	if err != nil {
		t.Fatalf("parseClientConfig() failed: %v", err)
	}

	if cfg.Vars["HOSTNAME"] != "override.example.com" {
		t.Errorf("expected HOSTNAME='override.example.com', got '%s'", cfg.Vars["HOSTNAME"])
	}
}

func TestLoadClientConfigs_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two config files
	configData := map[string]string{
		"mozilla.json": `{
			"vendor": "Mozilla",
			"templateFile": "/templates/mozilla.xml.tmpl",
			"supportedEndpoints": ["/endpoint-0"]
		}`,
		"microsoft.json": `{
			"vendor": "Microsoft",
			"templateFile": "/templates/microsoft.xml.tmpl",
			"supportedEndpoints": ["/endpoint-1"]
		}`,
	}

	for name, content := range configData {
		configPath := filepath.Join(tmpDir, name)
		err := os.WriteFile(configPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test config %s: %v", name, err)
		}
	}

	configs, err := LoadClientConfigs(tmpDir)
	if err != nil {
		t.Fatalf("LoadClientConfigs() failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	// Verify both vendors are loaded
	vendors := make(map[string]bool)
	for _, cfg := range configs {
		vendors[cfg.Vendor] = true
	}

	if !vendors["MOZILLA"] {
		t.Errorf("expected MOZILLA vendor to be loaded")
	}

	if !vendors["MICROSOFT"] {
		t.Errorf("expected MICROSOFT vendor to be loaded")
	}
}

func TestLoadClientConfigs_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "bad.json")
	err := os.WriteFile(configPath, []byte("{ invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err = LoadClientConfigs(tmpDir)
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}
