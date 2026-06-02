package config

import (
	"os"
	"testing"
)

func TestLoad_EnvVarMode(t *testing.T) {
	// Setup: set environment variables
	os.Setenv("SUPPORTED_DOMAINS", "example.com,test.org")
	os.Setenv("LISTEN_ADDR", ":9000")
	os.Setenv("ISDBENABLED", "false")
	defer func() {
		os.Unsetenv("SUPPORTED_DOMAINS")
		os.Unsetenv("LISTEN_ADDR")
		os.Unsetenv("ISDBENABLED")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(cfg.SupportedDomains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(cfg.SupportedDomains))
	}

	if cfg.SupportedDomains[0] != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", cfg.SupportedDomains[0])
	}

	if cfg.SupportedDomains[1] != "test.org" {
		t.Errorf("expected 'test.org', got '%s'", cfg.SupportedDomains[1])
	}

	if cfg.ListenAddr != ":9000" {
		t.Errorf("expected ':9000', got '%s'", cfg.ListenAddr)
	}

	if cfg.IsDBEnabled {
		t.Errorf("expected IsDBEnabled=false, got true")
	}
}

func TestLoad_DomainsCaseInsensitive(t *testing.T) {
	os.Setenv("SUPPORTED_DOMAINS", "EXAMPLE.COM,Test.Org")
	os.Setenv("ISDBENABLED", "false")
	defer func() {
		os.Unsetenv("SUPPORTED_DOMAINS")
		os.Unsetenv("ISDBENABLED")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SupportedDomains[0] != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", cfg.SupportedDomains[0])
	}

	if cfg.SupportedDomains[1] != "test.org" {
		t.Errorf("expected 'test.org', got '%s'", cfg.SupportedDomains[1])
	}
}

func TestLoad_DomainsWhitespaceHandling(t *testing.T) {
	os.Setenv("SUPPORTED_DOMAINS", "  example.com  , test.org,  ")
	os.Setenv("ISDBENABLED", "false")
	defer func() {
		os.Unsetenv("SUPPORTED_DOMAINS")
		os.Unsetenv("ISDBENABLED")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(cfg.SupportedDomains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(cfg.SupportedDomains))
	}
}

func TestLoad_ListenAddrDefault(t *testing.T) {
	os.Setenv("ISDBENABLED", "false")
	defer os.Unsetenv("ISDBENABLED")
	if addr, ok := os.LookupEnv("LISTEN_ADDR"); ok {
		os.Unsetenv("LISTEN_ADDR")
		defer os.Setenv("LISTEN_ADDR", addr)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default ':8080', got '%s'", cfg.ListenAddr)
	}
}

func TestLoad_DBEnabledMySQLDefaults(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBDRIVER", "mysql")
	os.Setenv("DBNAME", "testdb")
	os.Setenv("DBUSER", "testuser")
	os.Setenv("QUERY_DOMAINS", "SELECT domain FROM domain WHERE domain = ?")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBDRIVER")
		os.Unsetenv("DBNAME")
		os.Unsetenv("DBUSER")
		os.Unsetenv("QUERY_DOMAINS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if !cfg.IsDBEnabled {
		t.Errorf("expected IsDBEnabled=true, got false")
	}

	if cfg.DBPort != "3306" {
		t.Errorf("expected default port '3306' for MySQL, got '%s'", cfg.DBPort)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.DBHost)
	}
}

func TestLoad_DBEnabledPostgresDefaults(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBDRIVER", "postgres")
	os.Setenv("DBNAME", "testdb")
	os.Setenv("DBUSER", "testuser")
	os.Setenv("QUERY_DOMAINS", "SELECT domain FROM domain WHERE domain = $1")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBDRIVER")
		os.Unsetenv("DBNAME")
		os.Unsetenv("DBUSER")
		os.Unsetenv("QUERY_DOMAINS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DBPort != "5432" {
		t.Errorf("expected default port '5432' for Postgres, got '%s'", cfg.DBPort)
	}
}

func TestLoad_DBEnabledMissingQueryDomains(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBNAME", "testdb")
	os.Setenv("DBUSER", "testuser")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBNAME")
		os.Unsetenv("DBUSER")
	}()

	_, err := Load()
	if err == nil {
		t.Errorf("expected error for missing QUERY_DOMAINS, got nil")
	}
}

func TestLoad_DBEnabledMissingDBName(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBUSER", "testuser")
	os.Setenv("QUERY_DOMAINS", "SELECT domain FROM domain WHERE domain = ?")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBUSER")
		os.Unsetenv("QUERY_DOMAINS")
	}()

	_, err := Load()
	if err == nil {
		t.Errorf("expected error for missing DBNAME, got nil")
	}
}

func TestLoad_DBEnabledMissingDBUser(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBNAME", "testdb")
	os.Setenv("QUERY_DOMAINS", "SELECT domain FROM domain WHERE domain = ?")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBNAME")
		os.Unsetenv("QUERY_DOMAINS")
	}()

	_, err := Load()
	if err == nil {
		t.Errorf("expected error for missing DBUSER, got nil")
	}
}

func TestLoad_InvalidDBDriver(t *testing.T) {
	os.Setenv("ISDBENABLED", "true")
	os.Setenv("DBDRIVER", "invalid")
	os.Setenv("DBNAME", "testdb")
	os.Setenv("DBUSER", "testuser")
	os.Setenv("QUERY_DOMAINS", "SELECT domain FROM domain WHERE domain = ?")
	defer func() {
		os.Unsetenv("ISDBENABLED")
		os.Unsetenv("DBDRIVER")
		os.Unsetenv("DBNAME")
		os.Unsetenv("DBUSER")
		os.Unsetenv("QUERY_DOMAINS")
	}()

	_, err := Load()
	if err == nil {
		t.Errorf("expected error for invalid DBDRIVER, got nil")
	}
}

func TestValidatePlaceholders_MySQL(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"valid_single_placeholder", "SELECT * FROM domain WHERE domain = ?", false},
		{"valid_multiple_placeholders", "SELECT * FROM domain WHERE domain = ? OR domain = ?", false},
		{"missing_placeholder", "SELECT * FROM domain WHERE domain = 'test'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlaceholders("mysql", "TEST_QUERY", tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlaceholders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePlaceholders_Postgres(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"valid_single_placeholder", "SELECT * FROM domain WHERE domain = $1", false},
		{"valid_multiple_placeholders", "SELECT * FROM domain WHERE domain = $1 OR domain = $2", false},
		{"missing_placeholder", "SELECT * FROM domain WHERE domain = 'test'", true},
		{"mysql_style_placeholder", "SELECT * FROM domain WHERE domain = ?", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlaceholders("postgres", "TEST_QUERY", tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlaceholders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		defVal   string
		expected string
	}{
		{"env_set", "TEST_VAR_CUSTOM", "custom_value", "default", "custom_value"},
		{"env_empty", "TEST_VAR_EMPTY", "", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := envOrDefault(tt.envVar, tt.defVal)
			if result != tt.expected {
				t.Errorf("envOrDefault() = %s, expected %s", result, tt.expected)
			}
		})
	}
}
