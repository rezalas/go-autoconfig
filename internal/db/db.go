package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-autoconfig/internal/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps a *sql.DB and exposes domain lookup methods using
// caller-supplied prepared statement templates from Config.
type DB struct {
	conn         *sql.DB
	queryDomains string
}

// Open creates and validates a new DB connection based on cfg.
// It registers the correct driver for mysql/mariadb/postgres.
func Open(cfg *config.Config) (*DB, error) {
	dsn, driverName, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("opening db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	return &DB{
		conn:         conn,
		queryDomains: cfg.QueryDomains,
	}, nil
}

// Close releases the underlying connection pool.
func (d *DB) Close() error {
	return d.conn.Close()
}

// DomainExists returns true if the given domain is found by QueryDomains.
// The domain value is bound as a parameter — never interpolated into the query.
func (d *DB) DomainExists(ctx context.Context, domain string) (bool, error) {
	row := d.conn.QueryRowContext(ctx, d.queryDomains, domain)
	var found string
	err := row.Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("querying domains: %w", err)
	}
	return true, nil
}

// buildDSN constructs the driver name and DSN string from config.
func buildDSN(cfg *config.Config) (dsn, driverName string, err error) {
	switch cfg.DBDriver {
	case "mysql", "mariadb":
		// format: user:password@tcp(host:port)/dbname?parseTime=true
		driverName = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	case "postgres":
		// format: host=... port=... user=... password=... dbname=... sslmode=disable
		driverName = "pgx"
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=prefer",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	default:
		return "", "", fmt.Errorf("unsupported DBDRIVER: %q", cfg.DBDriver)
	}
	return dsn, driverName, nil
}
