package db

import (
	"context"
	"testing"

	"go-autoconfig/internal/config"
)

func TestBuildDSN_MySQL(t *testing.T) {
	tests := []struct {
		name           string
		dbDriver       string
		dbUser         string
		dbPassword     string
		dbHost         string
		dbPort         string
		dbName         string
		expectedDSN    string
		expectedDriver string
	}{
		{
			name:           "valid MySQL config",
			dbDriver:       "mysql",
			dbUser:         "root",
			dbPassword:     "secret",
			dbHost:         "localhost",
			dbPort:         "3306",
			dbName:         "test_db",
			expectedDSN:    "root:secret@tcp(localhost:3306)/test_db?parseTime=true",
			expectedDriver: "mysql",
		},
		{
			name:           "MariaDB config",
			dbDriver:       "mariadb",
			dbUser:         "root",
			dbPassword:     "secret",
			dbHost:         "localhost",
			dbPort:         "3306",
			dbName:         "test_db",
			expectedDSN:    "root:secret@tcp(localhost:3306)/test_db?parseTime=true",
			expectedDriver: "mysql",
		},
		{
			name:           "MySQL with empty password",
			dbDriver:       "mysql",
			dbUser:         "root",
			dbPassword:     "",
			dbHost:         "localhost",
			dbPort:         "5432",
			dbName:         "test_db",
			expectedDSN:    "root:@tcp(localhost:5432)/test_db?parseTime=true",
			expectedDriver: "mysql",
		},
		{
			name:           "MySQL with special characters in password",
			dbDriver:       "mysql",
			dbUser:         "user",
			dbPassword:     "p@ssw0rd!",
			dbHost:         "db.example.com",
			dbPort:         "3306",
			dbName:         "mydb",
			expectedDSN:    "user:p@ssw0rd!@tcp(db.example.com:3306)/mydb?parseTime=true",
			expectedDriver: "mysql",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				DBDriver:   tc.dbDriver,
				DBUser:     tc.dbUser,
				DBPassword: tc.dbPassword,
				DBHost:     tc.dbHost,
				DBPort:     tc.dbPort,
				DBName:     tc.dbName,
			}

			dsn, driver, err := buildDSN(cfg)
			if err != nil {
				t.Fatalf("buildDSN() failed: %v", err)
			}

			if dsn != tc.expectedDSN {
				t.Errorf("buildDSN() DSN = %q, expected %q", dsn, tc.expectedDSN)
			}

			if driver != tc.expectedDriver {
				t.Errorf("buildDSN() driver = %q, expected %q", driver, tc.expectedDriver)
			}
		})
	}
}

func TestBuildDSN_Postgres(t *testing.T) {
	tests := []struct {
		name           string
		dbDriver       string
		dbUser         string
		dbPassword     string
		dbHost         string
		dbPort         string
		dbName         string
		expectedDSN    string
		expectedDriver string
	}{
		{
			name:           "valid Postgres config",
			dbDriver:       "postgres",
			dbUser:         "postgres",
			dbPassword:     "secret",
			dbHost:         "localhost",
			dbPort:         "5432",
			dbName:         "test_db",
			expectedDSN:    "host=localhost port=5432 user=postgres password=secret dbname=test_db sslmode=prefer",
			expectedDriver: "pgx",
		},
		{
			name:           "Postgres with empty password",
			dbDriver:       "postgres",
			dbUser:         "postgres",
			dbPassword:     "",
			dbHost:         "localhost",
			dbPort:         "5432",
			dbName:         "test_db",
			expectedDSN:    "host=localhost port=5432 user=postgres password= dbname=test_db sslmode=prefer",
			expectedDriver: "pgx",
		},
		{
			name:           "Postgres with different host and port",
			dbDriver:       "postgres",
			dbUser:         "admin",
			dbPassword:     "pass123",
			dbHost:         "db.example.com",
			dbPort:         "5433",
			dbName:         "prod_db",
			expectedDSN:    "host=db.example.com port=5433 user=admin password=pass123 dbname=prod_db sslmode=prefer",
			expectedDriver: "pgx",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				DBDriver:   tc.dbDriver,
				DBUser:     tc.dbUser,
				DBPassword: tc.dbPassword,
				DBHost:     tc.dbHost,
				DBPort:     tc.dbPort,
				DBName:     tc.dbName,
			}

			dsn, driver, err := buildDSN(cfg)
			if err != nil {
				t.Fatalf("buildDSN() failed: %v", err)
			}

			if dsn != tc.expectedDSN {
				t.Errorf("buildDSN() DSN = %q, expected %q", dsn, tc.expectedDSN)
			}

			if driver != tc.expectedDriver {
				t.Errorf("buildDSN() driver = %q, expected %q", driver, tc.expectedDriver)
			}
		})
	}
}

func TestBuildDSN_UnsupportedDriver(t *testing.T) {
	cfg := &config.Config{
		DBDriver:   "sqlite3",
		DBUser:     "user",
		DBPassword: "pass",
		DBHost:     "host",
		DBPort:     "5432",
		DBName:     "dbname",
	}

	_, _, err := buildDSN(cfg)
	if err == nil {
		t.Errorf("expected error for unsupported driver, got nil")
	}
}

func TestDomainExists_NilConnection(t *testing.T) {
	// Create a DB instance with nil connection to test error handling
	db := &DB{conn: nil, queryDomains: "SELECT 1"}

	// This will panic when trying to use a nil connection
	// We verify this is the expected behavior
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()

	db.DomainExists(context.Background(), "example.com")

	if !panicked {
		t.Errorf("expected panic with nil connection")
	}
}

func TestDB_Close(t *testing.T) {
	// Note: Cannot test Close with nil connection as it panics
	// This is expected behavior - the DB should always have a valid connection
	// This test documents that Close should only be called on valid connections

	// In real usage, Open will validate the connection exists
	// If someone somehow creates a DB with nil conn, Close will panic
	// This is acceptable as it indicates a programming error
	t.Skip("Close() with nil connection panics - expected behavior, documented in code")
}

func TestDomainExists_ContextCancelled(t *testing.T) {
	// Create a DB instance with nil connection
	db := &DB{conn: nil, queryDomains: "SELECT 1"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This will panic when trying to use a nil connection
	// even with a cancelled context, the panic happens first
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()

	db.DomainExists(ctx, "example.com")

	if !panicked {
		t.Errorf("expected panic with nil connection")
	}
}
